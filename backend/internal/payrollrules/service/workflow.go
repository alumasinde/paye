package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Workflow struct{ DB *sql.DB }

func (s Workflow) SubmitReview(ctx context.Context, ruleSetID, adminID uint64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='IN_REVIEW' WHERE id=? AND status='DRAFT'`, ruleSetID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("only draft rule sets can be submitted")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO payroll_rule_change_requests(public_id,rule_set_id,requested_by,status) VALUES(UUID(),?,?, 'OPEN')`, ruleSetID, adminID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Workflow) Review(ctx context.Context, ruleSetID, reviewerID uint64, approve bool, comment string) error {
	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `UPDATE payroll_rule_change_requests SET status=?,reviewed_by=?,review_comment=?,reviewed_at=UTC_TIMESTAMP() WHERE rule_set_id=? AND status='OPEN'`, status, reviewerID, comment, ruleSetID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return errors.New("no open review request")
	}
	if approve {
		_, err = tx.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='IN_REVIEW' WHERE id=?`, ruleSetID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='DRAFT' WHERE id=?`, ruleSetID)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Publish flips a rule set to PUBLISHED and, in the same transaction,
// materializes its components into the live calculation schema
// (rule_definitions/rule_versions/rule_parameters/rule_bands) that
// internal/rules/repository.ResolvedApplicable and the calculator engine
// actually read. Without this, "publishing" a rule set only changed a
// status flag - nothing it contained ever reached the live calculator.
func (s Workflow) Publish(ctx context.Context, ruleSetID, publisherID uint64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var jurisdiction, code, versionCode string
	var versionNumber int
	var from time.Time
	var to sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT jurisdiction,effective_from,effective_to,code,version_number FROM payroll_rule_sets WHERE id=? AND status='IN_REVIEW'`, ruleSetID).
		Scan(&jurisdiction, &from, &to, &code, &versionNumber)
	if err != nil {
		return errors.New("rule set is not approved for publishing")
	}
	versionCode = fmt.Sprintf("%s_V%d", code, versionNumber)

	var approved int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payroll_rule_change_requests WHERE rule_set_id=? AND status='APPROVED'`, ruleSetID).Scan(&approved)
	if err != nil || approved < 1 {
		return errors.New("approved change request required")
	}
	var overlap int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payroll_rule_sets WHERE id<>? AND jurisdiction=? AND code=? AND status='PUBLISHED' AND effective_from<=? AND (effective_to IS NULL OR effective_to>=?)`, ruleSetID, jurisdiction, code, from, from).Scan(&overlap)
	if err != nil {
		return err
	}
	if overlap > 0 {
		return errors.New("published effective-date overlap detected")
	}

	if err := materializeComponents(ctx, tx, ruleSetID, versionCode, from, to, publisherID); err != nil {
		return fmt.Errorf("materialize live rules: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='PUBLISHED',published_at=UTC_TIMESTAMP(),published_by=? WHERE id=? AND status='IN_REVIEW'`, publisherID, ruleSetID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO payroll_rule_publish_events(public_id,rule_set_id,published_by) VALUES(UUID(),?,?)`, ruleSetID, publisherID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Workflow) Archive(ctx context.Context, ruleSetID uint64) error {
	_, err := s.DB.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='ARCHIVED' WHERE id=? AND status<>'ARCHIVED'`, ruleSetID)
	return err
}

type componentRow struct {
	Code                 string
	ComponentType        string
	Name                 string
	CalculationOrder     int
	ReducesTaxableIncome bool
	ReducesNetPay        bool
	FormulaType          string
	Payload              []byte
}

type fixedPayload struct {
	Amount float64 `json:"amount"`
}
type percentagePayload struct {
	Rate float64 `json:"rate"`
}
type percentageWithMinimumPayload struct {
	Rate          float64 `json:"rate"`
	MinimumAmount float64 `json:"minimum_amount"`
}
type cappedPercentagePayload struct {
	Rate                float64 `json:"rate"`
	UpperEarningsLimit  float64 `json:"upper_earnings_limit"`
	MaximumContribution float64 `json:"maximum_contribution"`
}
type bandsPayload struct {
	Bands []struct {
		From        float64  `json:"from"`
		To          *float64 `json:"to,omitempty"`
		Rate        float64  `json:"rate"`
		FixedAmount float64  `json:"fixed_amount,omitempty"`
	} `json:"bands"`
}

var componentTypeToCategory = map[string]string{
	"PAYE_BANDS":          "INCOME_TAX",
	"RELIEF":              "RELIEF",
	"STATUTORY_DEDUCTION": "STATUTORY_DEDUCTION",
}
var formulaTypeToMethod = map[string]string{
	"FIXED":                   "FIXED_AMOUNT",
	"PERCENTAGE":              "PERCENTAGE",
	"PERCENTAGE_WITH_MINIMUM": "PERCENTAGE_WITH_MINIMUM",
	"CAPPED_PERCENTAGE":       "CAPPED_PERCENTAGE",
	"BANDS":                   "PROGRESSIVE_BANDS",
	"TIERED_FIXED_AMOUNT":     "TIERED_FIXED_AMOUNT",
}

func materializeComponents(ctx context.Context, tx *sql.Tx, ruleSetID uint64, versionCode string, from time.Time, to sql.NullTime, publisherID uint64) error {
	rows, err := tx.QueryContext(ctx, `SELECT component_code,component_type,name,calculation_order,reduces_taxable_income,reduces_net_pay,formula_type,payload
		FROM payroll_rule_components WHERE rule_set_id=? AND is_active=1`, ruleSetID)
	if err != nil {
		return err
	}
	var components []componentRow
	for rows.Next() {
		var c componentRow
		if err := rows.Scan(&c.Code, &c.ComponentType, &c.Name, &c.CalculationOrder, &c.ReducesTaxableIncome, &c.ReducesNetPay, &c.FormulaType, &c.Payload); err != nil {
			rows.Close()
			return err
		}
		components = append(components, c)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	for _, c := range components {
		if c.ComponentType == "CONFIGURATION" {
			continue // metadata only, not a calculation rule - nothing to materialize
		}
		category, ok := componentTypeToCategory[c.ComponentType]
		if !ok {
			return fmt.Errorf("%s: unknown component type %q", c.Code, c.ComponentType)
		}
		methodCode, ok := formulaTypeToMethod[c.FormulaType]
		if !ok {
			return fmt.Errorf("%s: formula type %q cannot be published - use Fixed amount, Percentage, Percentage with minimum, Capped percentage, Progressive bands, or Tiered fixed amount", c.Code, c.FormulaType)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO rule_definitions(code,name,category) VALUES(?,?,?)
			ON DUPLICATE KEY UPDATE name=VALUES(name), category=VALUES(category)`, c.Code, c.Name, category); err != nil {
			return fmt.Errorf("%s: upsert rule_definitions: %w", c.Code, err)
		}
		var definitionID, methodID uint64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM rule_definitions WHERE code=?`, c.Code).Scan(&definitionID); err != nil {
			return fmt.Errorf("%s: %w", c.Code, err)
		}
		if err := tx.QueryRowContext(ctx, `SELECT id FROM calculation_methods WHERE code=?`, methodCode).Scan(&methodID); err != nil {
			return fmt.Errorf("%s: calculation method %q not found: %w", c.Code, methodCode, err)
		}

		res, err := tx.ExecContext(ctx, `INSERT INTO rule_versions
			(rule_definition_id,calculation_method_id,version_code,version_name,status,effective_from,effective_to,calculation_order,affects_taxable_income,affects_net_pay,source_summary,published_at,published_by)
			VALUES(?,?,?,?,'PUBLISHED',?,?,?,?,?,?,UTC_TIMESTAMP(),?)`,
			definitionID, methodID, versionCode+"_"+c.Code, c.Name, from, to, c.CalculationOrder, c.ReducesTaxableIncome, c.ReducesNetPay, "Published via admin rule set workflow", publisherID)
		if err != nil {
			return fmt.Errorf("%s: insert rule_versions: %w", c.Code, err)
		}
		versionID, _ := res.LastInsertId()

		switch c.FormulaType {
		case "FIXED":
			var p fixedPayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid FIXED payload: %w", c.Code, err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO rule_parameters(rule_version_id,parameter_name,parameter_type,value_decimal) VALUES(?,?,?,?)`,
				versionID, "amount", "DECIMAL", p.Amount); err != nil {
				return fmt.Errorf("%s: insert rule_parameters: %w", c.Code, err)
			}
		case "PERCENTAGE":
			var p percentagePayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid PERCENTAGE payload: %w", c.Code, err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO rule_parameters(rule_version_id,parameter_name,parameter_type,value_decimal) VALUES(?,?,?,?)`,
				versionID, "rate", "DECIMAL", p.Rate); err != nil {
				return fmt.Errorf("%s: insert rule_parameters: %w", c.Code, err)
			}
		case "PERCENTAGE_WITH_MINIMUM":
			var p percentageWithMinimumPayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid PERCENTAGE_WITH_MINIMUM payload: %w", c.Code, err)
			}
			params := []struct {
				name string
				val  float64
			}{{"rate", p.Rate}, {"minimum_amount", p.MinimumAmount}}
			for _, pr := range params {
				if _, err := tx.ExecContext(ctx, `INSERT INTO rule_parameters(rule_version_id,parameter_name,parameter_type,value_decimal) VALUES(?,?,?,?)`,
					versionID, pr.name, "DECIMAL", pr.val); err != nil {
					return fmt.Errorf("%s: insert rule_parameters (%s): %w", c.Code, pr.name, err)
				}
			}
		case "CAPPED_PERCENTAGE":
			var p cappedPercentagePayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid CAPPED_PERCENTAGE payload: %w", c.Code, err)
			}
			params := []struct {
				name string
				val  float64
			}{{"rate", p.Rate}, {"upper_earnings_limit", p.UpperEarningsLimit}, {"maximum_contribution", p.MaximumContribution}}
			for _, pr := range params {
				if _, err := tx.ExecContext(ctx, `INSERT INTO rule_parameters(rule_version_id,parameter_name,parameter_type,value_decimal) VALUES(?,?,?,?)`,
					versionID, pr.name, "DECIMAL", pr.val); err != nil {
					return fmt.Errorf("%s: insert rule_parameters (%s): %w", c.Code, pr.name, err)
				}
			}
		case "BANDS":
			var p bandsPayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid BANDS payload: %w", c.Code, err)
			}
			if len(p.Bands) == 0 {
				return fmt.Errorf("%s: BANDS payload has no bands", c.Code)
			}
			for i, b := range p.Bands {
				if _, err := tx.ExecContext(ctx, `INSERT INTO rule_bands(rule_version_id,from_amount,to_amount,rate,display_order) VALUES(?,?,?,?,?)`,
					versionID, b.From, b.To, b.Rate, i+1); err != nil {
					return fmt.Errorf("%s: insert rule_bands: %w", c.Code, err)
				}
			}
		case "TIERED_FIXED_AMOUNT":
			var p bandsPayload
			if err := json.Unmarshal(c.Payload, &p); err != nil {
				return fmt.Errorf("%s: invalid TIERED_FIXED_AMOUNT payload: %w", c.Code, err)
			}
			if len(p.Bands) == 0 {
				return fmt.Errorf("%s: TIERED_FIXED_AMOUNT payload has no bands", c.Code)
			}
			for i, b := range p.Bands {
				if _, err := tx.ExecContext(ctx, `INSERT INTO rule_bands(rule_version_id,from_amount,to_amount,fixed_amount,display_order) VALUES(?,?,?,?,?)`,
					versionID, b.From, b.To, b.FixedAmount, i+1); err != nil {
					return fmt.Errorf("%s: insert rule_bands: %w", c.Code, err)
				}
			}
		}
	}
	return nil
}
