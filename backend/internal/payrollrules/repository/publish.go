package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
)

// PublishToLive is the single bridge between the admin workspace and the
// calculator's immutable rule_versions catalog. It never edits a published
// live version; every published admin component creates one new version.
func (r Repository) PublishToLive(ctx context.Context, publicID string, publisherID uint64) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var rsID uint64
	var code, jurisdiction, status string
	var from time.Time
	var to sql.NullTime
	if err = tx.QueryRowContext(ctx, `SELECT id,code,jurisdiction,status,effective_from,effective_to FROM payroll_rule_sets WHERE public_id=? FOR UPDATE`, publicID).Scan(&rsID, &code, &jurisdiction, &status, &from, &to); err != nil {
		return fmt.Errorf("load rule set: %w", err)
	}
	if jurisdiction != "KE" {
		return fmt.Errorf("unsupported jurisdiction %q", jurisdiction)
	}
	if status != "IN_REVIEW" {
		return fmt.Errorf("rule set is not approved for publishing")
	}
	var approved int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payroll_rule_change_requests WHERE rule_set_id=? AND status='APPROVED'`, rsID).Scan(&approved); err != nil || approved == 0 {
		return fmt.Errorf("approved change request required")
	}
	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payroll_rule_set_live_versions WHERE rule_set_id=?`, rsID).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return fmt.Errorf("rule set already has live versions")
	}

	rows, err := tx.QueryContext(ctx, `SELECT component_code,component_type,name,calculation_order,reduces_taxable_income,reduces_net_pay,formula_type,payload,is_active FROM payroll_rule_components WHERE rule_set_id=? AND is_active=1 ORDER BY calculation_order,id`, rsID)
	if err != nil {
		return err
	}
	defer rows.Close()
	var components []model.Component
	for rows.Next() {
		var c model.Component
		var payload []byte
		if err := rows.Scan(&c.ComponentCode, &c.ComponentType, &c.Name, &c.CalculationOrder, &c.ReducesTaxableIncome, &c.ReducesNetPay, &c.FormulaType, &payload, &c.IsActive); err != nil {
			return err
		}
		if err := json.Unmarshal(payload, &c.Payload); err != nil {
			return fmt.Errorf("decode component %s: %w", c.ComponentCode, err)
		}
		components = append(components, c)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(components) == 0 {
		return fmt.Errorf("rule set has no active components")
	}

	for _, c := range components {
		if err := publishComponent(ctx, tx, rsID, code, from, to, publisherID, c); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE payroll_rule_sets SET status='PUBLISHED',published_at=UTC_TIMESTAMP(),published_by=? WHERE id=? AND status='IN_REVIEW'`, publisherID, rsID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO payroll_rule_publish_events(public_id,rule_set_id,published_by) VALUES(UUID(),?,?)`, rsID, publisherID); err != nil {
		return err
	}
	return tx.Commit()
}

func publishComponent(ctx context.Context, tx *sql.Tx, ruleSetID uint64, setCode string, from time.Time, to sql.NullTime, publisherID uint64, c model.Component) error {
	var definitionID, methodID uint64
	code := strings.ToUpper(strings.TrimSpace(c.ComponentCode))
	if code == "" {
		return fmt.Errorf("component code is required")
	}
	if err := tx.QueryRowContext(ctx, `SELECT id FROM rule_definitions WHERE code=? AND is_active=1`, code).Scan(&definitionID); err != nil {
		return fmt.Errorf("component %s must match an active live rule definition: %w", code, err)
	}
	method := methodCode(c)
	if err := tx.QueryRowContext(ctx, `SELECT id FROM calculation_methods WHERE code=? AND is_active=1`, method).Scan(&methodID); err != nil {
		return fmt.Errorf("component %s uses unsupported calculation method %s", code, method)
	}
	// Close the immediately preceding live version at the day before this
	// version starts. This preserves historical calculations while ensuring the
	// calculator can resolve exactly one version for every date. A future
	// published version or an invalid backdated overlap is rejected.
	var previousID uint64
	var previousFrom time.Time
	prevErr := tx.QueryRowContext(ctx, `SELECT id,effective_from FROM rule_versions WHERE rule_definition_id=? AND status='PUBLISHED' AND effective_from < ? AND (effective_to IS NULL OR effective_to >= ?) ORDER BY effective_from DESC LIMIT 1 FOR UPDATE`, definitionID, from, from).Scan(&previousID, &previousFrom)
	if prevErr != nil && prevErr != sql.ErrNoRows {
		return prevErr
	}
	var conflicts int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM rule_versions WHERE rule_definition_id=? AND status='PUBLISHED' AND effective_from>=?`, definitionID, from).Scan(&conflicts); err != nil {
		return err
	}
	if conflicts > 0 {
		return fmt.Errorf("live effective-date overlap or future version for component %s", code)
	}
	if prevErr == nil {
		closeDate := from.AddDate(0, 0, -1)
		if closeDate.Before(previousFrom) {
			return fmt.Errorf("invalid effective date for component %s", code)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE rule_versions SET effective_to=? WHERE id=? AND status='PUBLISHED'`, closeDate, previousID); err != nil {
			return err
		}
	}
	versionCode := fmt.Sprintf("%s_ADMIN_%d_%s", code, ruleSetID, from.Format("20060102"))
	versionName := c.Name
	if versionName == "" {
		versionName = code + " published from admin rule set"
	}
	var liveID uint64
	res, err := tx.ExecContext(ctx, `INSERT INTO rule_versions(rule_definition_id,calculation_method_id,version_code,version_name,status,effective_from,effective_to,calculation_order,affects_taxable_income,affects_net_pay,source_summary,notes,published_at) VALUES(?,?,?,?,'PUBLISHED',?,?,?,?,?,?,?,UTC_TIMESTAMP())`, definitionID, methodID, versionCode, versionName, from, nullTimeValue(to), c.CalculationOrder, c.ReducesTaxableIncome, c.ReducesNetPay, "Published from Budget254 PAYE Admin", setCode)
	if err != nil {
		return fmt.Errorf("create live version for %s: %w", code, err)
	}
	id, _ := res.LastInsertId()
	liveID = uint64(id)
	if err := insertPayload(ctx, tx, liveID, c); err != nil {
		return fmt.Errorf("copy %s data: %w", code, err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO payroll_rule_set_live_versions(rule_set_id,rule_version_id,component_code) VALUES(?,?,?)`, ruleSetID, liveID, code); err != nil {
		return err
	}
	return nil
}
func nullTimeValue(v sql.NullTime) any {
	if v.Valid {
		return v.Time
	}
	return nil
}
func methodCode(c model.Component) string {
	switch c.FormulaType {
	case "FIXED":
		return "FIXED_AMOUNT"
	case "PERCENTAGE":
		return "PERCENTAGE"
	case "BANDS":
		if c.ComponentType == "PAYE_BANDS" {
			return "PROGRESSIVE_BANDS"
		}
		if bands, ok := c.Payload["bands"].([]any); ok && len(bands) > 0 {
			if m, ok := bands[0].(map[string]any); ok {
				if _, ok := m["fixed_amount"]; ok {
					return "TIERED_FIXED_AMOUNT"
				}
			}
		}
		return "PROGRESSIVE_BANDS"
	default:
		return ""
	}
}
func insertPayload(ctx context.Context, tx *sql.Tx, versionID uint64, c model.Component) error {
	if c.FormulaType == "BANDS" {
		raw, ok := c.Payload["bands"].([]any)
		if !ok || len(raw) == 0 {
			return fmt.Errorf("bands payload is required")
		}
		for i, v := range raw {
			m, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid band %d", i+1)
			}
			from, ok := numberString(m["lower"])
			if !ok {
				from, ok = numberString(m["from"])
			}
			if !ok {
				return fmt.Errorf("band %d lower/from is required", i+1)
			}
			var upper, rate, fixed any
			if s, ok := numberString(m["upper"]); ok {
				upper = s
			}
			if s, ok := numberString(m["to"]); ok {
				upper = s
			}
			if s, ok := numberString(m["rate"]); ok {
				rate = s
			}
			if s, ok := numberString(m["fixed_amount"]); ok {
				fixed = s
			}
			if rate == nil && fixed == nil {
				return fmt.Errorf("band %d requires rate or fixed_amount", i+1)
			}
			label, _ := m["label"].(string)
			if _, err := tx.ExecContext(ctx, `INSERT INTO rule_bands(rule_version_id,from_amount,to_amount,rate,fixed_amount,display_order,label) VALUES(?,?,?,?,?,?,?)`, versionID, from, upper, rate, fixed, i+1, label); err != nil {
				return err
			}
		}
		return nil
	}
	for name, val := range c.Payload {
		if name == "bands" {
			continue
		}
		typ, dec, integer, boolean, text, ok := parameterValue(val)
		if !ok {
			return fmt.Errorf("unsupported payload parameter %s", name)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO rule_parameters(rule_version_id,parameter_name,parameter_type,value_decimal,value_integer,value_boolean,value_text) VALUES(?,?,?,?,?,?,?)`, versionID, name, typ, dec, integer, boolean, text); err != nil {
			return err
		}
	}
	return nil
}
func numberString(v any) (string, bool) {
	switch n := v.(type) {
	case float64:
		return fmt.Sprintf("%v", n), true
	case json.Number:
		return n.String(), true
	case string:
		if strings.TrimSpace(n) != "" {
			return n, true
		}
	}
	return "", false
}
func parameterValue(v any) (string, any, any, any, any, bool) {
	switch x := v.(type) {
	case bool:
		return "BOOLEAN", nil, nil, x, nil, true
	case float64:
		return "DECIMAL", fmt.Sprintf("%v", x), nil, nil, nil, true
	case json.Number:
		return "DECIMAL", x.String(), nil, nil, nil, true
	case string:
		return "TEXT", nil, nil, nil, x, true
	}
	return "", nil, nil, nil, nil, false
}

// InternalID resolves the public UUID used by the admin API to its database key.
func (r Repository) InternalID(ctx context.Context, publicID string) (uint64, error) {
	var id uint64
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM payroll_rule_sets WHERE public_id=?`, publicID).Scan(&id)
	return id, err
}
