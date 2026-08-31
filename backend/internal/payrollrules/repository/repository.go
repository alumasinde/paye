package repository

import (
    "context"
    "database/sql"
    "encoding/json"
    "github.com/google/uuid"
    "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
)

type Repository struct { DB *sql.DB }

func (r Repository) List(ctx context.Context) ([]model.RuleSet,error) {
    rows,err:=r.DB.QueryContext(ctx,`SELECT public_id,code,name,jurisdiction,DATE_FORMAT(effective_from,'%Y-%m-%d'),effective_to,status,version_number,COALESCE(source_reference,''),COALESCE(source_notes,''),created_at,updated_at FROM payroll_rule_sets ORDER BY effective_from DESC,version_number DESC`)
    if err!=nil{return nil,err}; defer rows.Close()
    out:=[]model.RuleSet{}
    for rows.Next(){var x model.RuleSet; var to sql.NullTime
        if err:=rows.Scan(&x.PublicID,&x.Code,&x.Name,&x.Jurisdiction,&x.EffectiveFrom,&to,&x.Status,&x.VersionNumber,&x.SourceReference,&x.SourceNotes,&x.CreatedAt,&x.UpdatedAt);err!=nil{return nil,err}
        if to.Valid {v:=to.Time.Format("2006-01-02");x.EffectiveTo=&v};out=append(out,x)}
    return out,rows.Err()
}

func (r Repository) Create(ctx context.Context, x model.RuleSet, adminID uint64) (model.RuleSet,error) {    x.PublicID=uuid.NewString()
    if x.Jurisdiction=="" {x.Jurisdiction="KE"}; if x.Status=="" {x.Status="DRAFT"}
    tx,err:=r.DB.BeginTx(ctx,nil);if err!=nil{return x,err};defer tx.Rollback()
    var id uint64
    err=tx.QueryRowContext(ctx,`SELECT COALESCE(MAX(version_number),0)+1 FROM payroll_rule_sets WHERE code=?`,x.Code).Scan(&x.VersionNumber);if err!=nil{return x,err}
    res,err:=tx.ExecContext(ctx,`INSERT INTO payroll_rule_sets(public_id,code,name,jurisdiction,effective_from,effective_to,status,version_number,source_reference,source_notes,created_by) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,x.PublicID,x.Code,x.Name,x.Jurisdiction,x.EffectiveFrom,x.EffectiveTo,x.Status,x.VersionNumber,x.SourceReference,x.SourceNotes,adminID);if err!=nil{return x,err}
    id64,_:=res.LastInsertId();id=uint64(id64)
    for _,c:=range x.Components {c.PublicID=uuid.NewString(); payload,_:=json.Marshal(c.Payload);_,err=tx.ExecContext(ctx,`INSERT INTO payroll_rule_components(public_id,rule_set_id,component_code,component_type,name,calculation_order,reduces_taxable_income,reduces_net_pay,formula_type,payload,is_active) VALUES(?,?,?,?,?,?,?,?,?,?,?)`,c.PublicID,id,c.ComponentCode,c.ComponentType,c.Name,c.CalculationOrder,c.ReducesTaxableIncome,c.ReducesNetPay,c.FormulaType,payload,c.IsActive);if err!=nil{return x,err}}
    err=tx.Commit();return x,err
}

// IDByPublicID resolves a rule set's public (UUID) ID to its numeric
// database ID - the frontend only ever sees the public ID, but the
// workflow service (submit-review/approve/reject/publish/archive)
// operates on the numeric one.
func (r Repository) IDByPublicID(ctx context.Context, publicID string) (uint64, error) {
	var id uint64
	err := r.DB.QueryRowContext(ctx, `SELECT id FROM payroll_rule_sets WHERE public_id=?`, publicID).Scan(&id)
	return id, err
}

// GetByID returns a single rule set with its components, for the rule
// editor to load an existing draft.
func (r Repository) GetByID(ctx context.Context, id uint64) (model.RuleSet, error) {
	var x model.RuleSet
	var to sql.NullTime
	err := r.DB.QueryRowContext(ctx, `SELECT public_id,code,name,jurisdiction,DATE_FORMAT(effective_from,'%Y-%m-%d'),effective_to,status,version_number,COALESCE(source_reference,''),COALESCE(source_notes,''),created_at,updated_at FROM payroll_rule_sets WHERE id=?`, id).
		Scan(&x.PublicID, &x.Code, &x.Name, &x.Jurisdiction, &x.EffectiveFrom, &to, &x.Status, &x.VersionNumber, &x.SourceReference, &x.SourceNotes, &x.CreatedAt, &x.UpdatedAt)
	if err != nil {
		return x, err
	}
	if to.Valid {
		v := to.Time.Format("2006-01-02")
		x.EffectiveTo = &v
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT public_id,component_code,component_type,name,calculation_order,reduces_taxable_income,reduces_net_pay,formula_type,payload,is_active FROM payroll_rule_components WHERE rule_set_id=? ORDER BY calculation_order`, id)
	if err != nil {
		return x, err
	}
	defer rows.Close()
	for rows.Next() {
		var c model.Component
		var payload []byte
		if err := rows.Scan(&c.PublicID, &c.ComponentCode, &c.ComponentType, &c.Name, &c.CalculationOrder, &c.ReducesTaxableIncome, &c.ReducesNetPay, &c.FormulaType, &payload, &c.IsActive); err != nil {
			return x, err
		}
		_ = json.Unmarshal(payload, &c.Payload)
		x.Components = append(x.Components, c)
	}
	return x, rows.Err()
}
