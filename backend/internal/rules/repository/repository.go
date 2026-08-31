package repository
import (
 "context"; "database/sql"; "fmt"; "strings"; "time"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/model"
 rulesmodel "github.com/alumasinde/budget254-paye-api/internal/rules/model"
)
type Repository struct{ DB *sql.DB }
func (r Repository) Applicable(ctx context.Context, date time.Time) ([]model.Rule,error) {
 // Schema contract: rule_definitions, rule_versions, calculation_methods.
 // Additional parameters/bands are loaded by code-specific queries in the service layer.
 rows,err:=r.DB.QueryContext(ctx,`SELECT rd.code, rd.name, rv.version_code, cm.code, rv.effective_from, rv.effective_to
 FROM rule_definitions rd JOIN rule_versions rv ON rv.rule_definition_id=rd.id
 JOIN calculation_methods cm ON cm.id=rv.calculation_method_id
 WHERE rv.status='PUBLISHED' AND rv.effective_from<=? AND (rv.effective_to IS NULL OR rv.effective_to>=?)
 ORDER BY rd.code, rv.effective_from DESC`,date,date)
 if err!=nil{return nil,fmt.Errorf("query applicable rules: %w",err)}; defer rows.Close()
 out:=[]model.Rule{}; seen:=map[string]bool{}
 for rows.Next(){ var x model.Rule; var end sql.NullTime
  if err:=rows.Scan(&x.Code,&x.Name,&x.Version,&x.Method,&x.EffectiveFrom,&end);err!=nil{return nil,err}
  if seen[x.Code]{continue}; seen[x.Code]=true; if end.Valid{x.EffectiveTo=&end.Time}; out=append(out,x)
 }
 return out,rows.Err()
}

// ResolvedApplicable returns the published rule versions in effect on date,
// each with its calculation parameters (rule_parameters) and bands
// (rule_bands) fully loaded. This is what the calculation engine
// (internal/payroll/engine.Calculate) actually needs to compute an amount;
// Applicable above intentionally stays a lighter listing query used by the
// GET /api/v1/payroll/rules endpoint.
func (r Repository) ResolvedApplicable(ctx context.Context, date time.Time) ([]rulesmodel.ResolvedRule, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT rd.id, rd.code, rd.name, rd.category, rd.description,
		rv.id, rv.version_code, rv.version_name, cm.code,
		rv.calculation_order, rv.affects_taxable_income, rv.affects_net_pay,
		rv.effective_from, rv.effective_to
		FROM rule_definitions rd JOIN rule_versions rv ON rv.rule_definition_id = rd.id
		JOIN calculation_methods cm ON cm.id = rv.calculation_method_id
		WHERE rv.status = 'PUBLISHED' AND rv.effective_from <= ? AND (rv.effective_to IS NULL OR rv.effective_to >= ?)
		ORDER BY rd.code, rv.effective_from DESC`, date, date)
	if err != nil {
		return nil, fmt.Errorf("query resolved rules: %w", err)
	}
	defer rows.Close()

	type scanned struct {
		rule      rulesmodel.ResolvedRule
		versionID uint64
	}
	var collected []scanned
	seen := map[string]bool{}
	for rows.Next() {
		var x rulesmodel.ResolvedRule
		var description sql.NullString
		var end sql.NullTime
		var affectsTaxable, affectsNet bool
		if err := rows.Scan(&x.ID, &x.Code, &x.Name, &x.Category, &description,
			&x.VersionID, &x.VersionCode, &x.VersionName, &x.CalculationMethod,
			&x.CalculationOrder, &affectsTaxable, &affectsNet,
			&x.EffectiveFrom, &end); err != nil {
			return nil, err
		}
		if seen[x.Code] {
			continue
		}
		seen[x.Code] = true
		if description.Valid {
			d := description.String
			x.Description = &d
		}
		if end.Valid {
			t := end.Time
			x.EffectiveTo = &t
		}
		x.AffectsTaxableIncome = affectsTaxable
		x.AffectsNetPay = affectsNet
		collected = append(collected, scanned{rule: x, versionID: x.VersionID})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(collected) == 0 {
		return []rulesmodel.ResolvedRule{}, nil
	}

	out := make([]rulesmodel.ResolvedRule, len(collected))
	idx := make(map[uint64]int, len(collected))
	ids := make([]uint64, len(collected))
	for i, c := range collected {
		out[i] = c.rule
		idx[c.versionID] = i
		ids[i] = c.versionID
	}

	if err := r.loadParameters(ctx, ids, idx, out); err != nil {
		return nil, err
	}
	if err := r.loadBands(ctx, ids, idx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r Repository) loadParameters(ctx context.Context, ids []uint64, idx map[uint64]int, out []rulesmodel.ResolvedRule) error {
	placeholders, args := inArgs(ids)
	rows, err := r.DB.QueryContext(ctx, fmt.Sprintf(
		`SELECT rule_version_id, parameter_name, parameter_type, value_decimal, value_integer, value_boolean, value_text
		 FROM rule_parameters WHERE rule_version_id IN (%s)`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("query rule parameters: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var versionID uint64
		var name, ptype string
		var vDecimal, vText sql.NullString
		var vInteger sql.NullInt64
		var vBoolean sql.NullBool
		if err := rows.Scan(&versionID, &name, &ptype, &vDecimal, &vInteger, &vBoolean, &vText); err != nil {
			return err
		}
		i, ok := idx[versionID]
		if !ok {
			continue
		}
		p := rulesmodel.Parameter{Name: name, Type: ptype}
		if vDecimal.Valid {
			v := vDecimal.String
			p.Decimal = &v
		}
		if vInteger.Valid {
			v := vInteger.Int64
			p.Integer = &v
		}
		if vBoolean.Valid {
			v := vBoolean.Bool
			p.Boolean = &v
		}
		if vText.Valid {
			v := vText.String
			p.Text = &v
		}
		out[i].Parameters = append(out[i].Parameters, p)
	}
	return rows.Err()
}

func (r Repository) loadBands(ctx context.Context, ids []uint64, idx map[uint64]int, out []rulesmodel.ResolvedRule) error {
	placeholders, args := inArgs(ids)
	rows, err := r.DB.QueryContext(ctx, fmt.Sprintf(
		`SELECT rule_version_id, from_amount, to_amount, rate, fixed_amount, display_order, label
		 FROM rule_bands WHERE rule_version_id IN (%s) ORDER BY rule_version_id, display_order`, placeholders), args...)
	if err != nil {
		return fmt.Errorf("query rule bands: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var versionID uint64
		var from string
		var to, rate, fixed, label sql.NullString
		var order int
		if err := rows.Scan(&versionID, &from, &to, &rate, &fixed, &order, &label); err != nil {
			return err
		}
		i, ok := idx[versionID]
		if !ok {
			continue
		}
		b := rulesmodel.Band{From: from, Order: order}
		if to.Valid {
			v := to.String
			b.To = &v
		}
		if rate.Valid {
			v := rate.String
			b.Rate = &v
		}
		if fixed.Valid {
			v := fixed.String
			b.FixedAmount = &v
		}
		if label.Valid {
			v := label.String
			b.Label = &v
		}
		out[i].Bands = append(out[i].Bands, b)
	}
	return rows.Err()
}

func inArgs(ids []uint64) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return strings.Join(placeholders, ","), args
}
