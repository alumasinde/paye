package repository

import (
 "context"
 "database/sql"
 "errors"
 "encoding/json"
 "strings"
 "time"

 "github.com/google/uuid"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrun/model"
)

var (
 ErrNotFound = errors.New("not found")
 ErrConflict = errors.New("conflict")
)

type Repository struct{ DB *sql.DB }

func (r Repository) companyID(ctx context.Context, publicID string) (uint64,error) {
 var id uint64
 err:=r.DB.QueryRowContext(ctx,`SELECT id FROM companies WHERE public_id=? AND status='ACTIVE'`,publicID).Scan(&id)
 if errors.Is(err,sql.ErrNoRows){return 0,ErrNotFound}
 return id,err
}

func (r Repository) Create(ctx context.Context, companyPublicID, period string, start, end time.Time) (model.Detail, error) {
 cid, err := r.companyID(ctx, companyPublicID)
 if err != nil { return model.Detail{}, err }

 tx, err := r.DB.BeginTx(ctx, nil)
 if err != nil { return model.Detail{}, err }
 defer tx.Rollback()

 publicID := uuid.NewString()
 result, err := tx.ExecContext(ctx, `INSERT INTO payroll_runs(public_id,company_id,period_key,period_start,period_end,status)
 VALUES(?,?,?,?,?,'DRAFT')`, publicID, cid, period, start, end)
 if err != nil {
  if strings.Contains(strings.ToLower(err.Error()), "duplicate") { return model.Detail{}, ErrConflict }
  return model.Detail{}, err
 }

 runID, err := result.LastInsertId()
 if err != nil { return model.Detail{}, err }

 type snapshot struct {
  employeeID uint64
  number, first, middle, last, salary, frequency string
 }
 employees := make([]snapshot, 0)

 rows, err := tx.QueryContext(ctx, `SELECT e.id,e.employee_number,e.first_name,COALESCE(e.middle_name,''),e.last_name,
 CAST(s.basic_salary AS CHAR),s.pay_frequency
 FROM employees e
 JOIN employee_salary_history s ON s.id=(
   SELECT es.id FROM employee_salary_history es
   WHERE es.employee_id=e.id AND es.effective_from<=?
   AND (es.effective_to IS NULL OR es.effective_to>=?)
   ORDER BY es.effective_from DESC LIMIT 1
 )
 WHERE e.company_id=? AND e.status='ACTIVE' AND e.employment_date<=?
 AND (e.termination_date IS NULL OR e.termination_date>=?)
 ORDER BY e.first_name,e.last_name`, end, end, cid, end, start)
 if err != nil { return model.Detail{}, err }

 for rows.Next() {
  var item snapshot
  if err := rows.Scan(&item.employeeID, &item.number, &item.first, &item.middle, &item.last, &item.salary, &item.frequency); err != nil {
   rows.Close()
   return model.Detail{}, err
  }
  employees = append(employees, item)
 }
 if err := rows.Err(); err != nil {
  rows.Close()
  return model.Detail{}, err
 }
 if err := rows.Close(); err != nil { return model.Detail{}, err }

 for _, employee := range employees {
  _, err = tx.ExecContext(ctx, `INSERT INTO payroll_run_employees(public_id,payroll_run_id,employee_id,employee_number,first_name,middle_name,last_name,basic_salary,pay_frequency,status)
  VALUES(?,?,?,?,?,?,?,?,?,'PENDING')`,
   uuid.NewString(), runID, employee.employeeID, employee.number, employee.first, nilIf(employee.middle),
   employee.last, employee.salary, employee.frequency,
  )
  if err != nil { return model.Detail{}, err }
 }

 if err := tx.Commit(); err != nil { return model.Detail{}, err }
 return r.Get(ctx, companyPublicID, publicID)
}
func (r Repository) List(ctx context.Context, companyPublicID string)([]model.PayrollRun,error){
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return nil,err}
 rows,err:=r.DB.QueryContext(ctx,`SELECT pr.public_id,pr.period_key,pr.period_start,pr.period_end,pr.status,COUNT(pre.id),pr.created_at,pr.updated_at
 FROM payroll_runs pr LEFT JOIN payroll_run_employees pre ON pre.payroll_run_id=pr.id
 WHERE pr.company_id=? GROUP BY pr.id ORDER BY pr.period_start DESC,pr.created_at DESC`,cid)
 if err!=nil{return nil,err};defer rows.Close()
 out:=[]model.PayrollRun{}
 for rows.Next(){var v model.PayrollRun;if err:=rows.Scan(&v.PublicID,&v.Period,&v.PeriodStart,&v.PeriodEnd,&v.Status,&v.EmployeeCount,&v.CreatedAt,&v.UpdatedAt);err!=nil{return nil,err};out=append(out,v)}
 return out,rows.Err()
}

func (r Repository) Get(ctx context.Context, companyPublicID, runPublicID string)(model.Detail,error){
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return model.Detail{},err}
 var id uint64; var d model.Detail
 err=r.DB.QueryRowContext(ctx,`SELECT id,public_id,period_key,period_start,period_end,status,created_at,updated_at FROM payroll_runs WHERE company_id=? AND public_id=?`,cid,runPublicID).Scan(&id,&d.PublicID,&d.Period,&d.PeriodStart,&d.PeriodEnd,&d.Status,&d.CreatedAt,&d.UpdatedAt)
 if errors.Is(err,sql.ErrNoRows){return model.Detail{},ErrNotFound};if err!=nil{return model.Detail{},err}
 rows,err:=r.DB.QueryContext(ctx,`SELECT pre.public_id,e.public_id,pre.employee_number,pre.first_name,COALESCE(pre.middle_name,''),pre.last_name,
 CAST(pre.basic_salary AS CHAR),pre.pay_frequency,pre.status,
 COALESCE(CAST(pre.gross_salary AS CHAR),''),COALESCE(CAST(pre.taxable_income AS CHAR),''),
 COALESCE(CAST(pre.paye_before_relief AS CHAR),''),COALESCE(CAST(pre.relief AS CHAR),''),
 COALESCE(CAST(pre.paye AS CHAR),''),COALESCE(CAST(pre.statutory_deductions AS CHAR),''),
 COALESCE(CAST(pre.custom_deductions AS CHAR),''),COALESCE(CAST(pre.total_deductions AS CHAR),''),
 COALESCE(CAST(pre.net_salary AS CHAR),''),pre.rule_versions,pre.calculated_at,COALESCE(pre.error_message,'')
 FROM payroll_run_employees pre JOIN employees e ON e.id=pre.employee_id WHERE pre.payroll_run_id=? ORDER BY pre.first_name,pre.last_name`,id)
 if err!=nil{return model.Detail{},err};defer rows.Close()
 for rows.Next(){
  var e model.PayrollRunEmployee; var rules sql.NullString; var calculated sql.NullTime
  if err:=rows.Scan(&e.PublicID,&e.EmployeeID,&e.EmployeeNumber,&e.FirstName,&e.MiddleName,&e.LastName,&e.BasicSalary,&e.PayFrequency,&e.Status,
   &e.GrossSalary,&e.TaxableIncome,&e.PAYEBeforeRelief,&e.Relief,&e.PAYE,&e.StatutoryDeductions,&e.CustomDeductions,&e.TotalDeductions,&e.NetSalary,&rules,&calculated,&e.ErrorMessage);err!=nil{return model.Detail{},err}
  if calculated.Valid { t:=calculated.Time; e.CalculatedAt=&t }
  if rules.Valid && strings.TrimSpace(rules.String)!="" {
   e.RuleVersions=map[string]string{}
   _=json.Unmarshal([]byte(rules.String),&e.RuleVersions)
  }
  d.Employees=append(d.Employees,e)
 }
 if err:=rows.Err();err!=nil{return model.Detail{},err}
 d.EmployeeCount=len(d.Employees)

 workflowRows,err:=r.DB.QueryContext(ctx,`SELECT e.action,e.from_status,e.to_status,
 COALESCE(NULLIF(TRIM(CONCAT(u.first_name,' ',u.last_name)),''),u.email,''),e.created_at
 FROM payroll_run_workflow_events e
 LEFT JOIN users u ON u.id=e.actor_user_id
 WHERE e.payroll_run_id=?
 ORDER BY e.created_at ASC,e.id ASC`,id)
 if err!=nil{return model.Detail{},err}
 defer workflowRows.Close()
 for workflowRows.Next(){
  var event model.WorkflowEvent
  if err:=workflowRows.Scan(&event.Action,&event.FromStatus,&event.ToStatus,&event.ActorName,&event.CreatedAt);err!=nil{return model.Detail{},err}
  d.Workflow=append(d.Workflow,event)
 }
 if err:=workflowRows.Err();err!=nil{return model.Detail{},err}
 return d,nil
}

func (r Repository) Pending(ctx context.Context, companyPublicID, runPublicID string)(model.PayrollRun,[]model.PayrollRunEmployee,error){
 d,err:=r.Get(ctx,companyPublicID,runPublicID);if err!=nil{return model.PayrollRun{},nil,err}
 if d.Status!="DRAFT" && d.Status!="CALCULATION_FAILED" {return model.PayrollRun{},nil,ErrConflict}
 pending:=make([]model.PayrollRunEmployee,0)
 for _,e:=range d.Employees {if e.Status=="PENDING" || e.Status=="FAILED" {pending=append(pending,e)}}
 return d.PayrollRun,pending,nil
}

func (r Repository) SaveCalculation(ctx context.Context, companyPublicID, runPublicID, employeeRunID string, values model.PayrollRunEmployee) error {
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return err}
 _,err=r.DB.ExecContext(ctx,`UPDATE payroll_run_employees pre
 JOIN payroll_runs pr ON pr.id=pre.payroll_run_id
 SET pre.status='CALCULATED',pre.gross_salary=?,pre.taxable_income=?,pre.paye_before_relief=?,pre.relief=?,pre.paye=?,
 pre.statutory_deductions=?,pre.custom_deductions=?,pre.total_deductions=?,pre.net_salary=?,pre.rule_versions=?,pre.calculated_at=UTC_TIMESTAMP(),pre.error_message=NULL
 WHERE pr.company_id=? AND pr.public_id=? AND pre.public_id=?`,
 values.GrossSalary,values.TaxableIncome,values.PAYEBeforeRelief,values.Relief,values.PAYE,values.StatutoryDeductions,values.CustomDeductions,values.TotalDeductions,values.NetSalary,mustJSON(values.RuleVersions),cid,runPublicID,employeeRunID)
 return err
}

func (r Repository) SaveFailure(ctx context.Context, companyPublicID, runPublicID, employeeRunID,message string) error {
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return err}
 _,err=r.DB.ExecContext(ctx,`UPDATE payroll_run_employees pre JOIN payroll_runs pr ON pr.id=pre.payroll_run_id
 SET pre.status='FAILED',pre.error_message=? WHERE pr.company_id=? AND pr.public_id=? AND pre.public_id=?`,message,cid,runPublicID,employeeRunID)
 return err
}

func (r Repository) FinalizeCalculation(ctx context.Context, companyPublicID, runPublicID, userPublicID string) (model.CalculationSummary,error) {
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return model.CalculationSummary{},err}
 tx,err:=r.DB.BeginTx(ctx,nil);if err!=nil{return model.CalculationSummary{},err};defer tx.Rollback()

 var id uint64; var out model.CalculationSummary
 err=tx.QueryRowContext(ctx,`SELECT id,public_id,period_key,period_start,period_end,status,created_at,updated_at FROM payroll_runs WHERE company_id=? AND public_id=? FOR UPDATE`,cid,runPublicID).
  Scan(&id,&out.PublicID,&out.Period,&out.PeriodStart,&out.PeriodEnd,&out.Status,&out.CreatedAt,&out.UpdatedAt)
 if errors.Is(err,sql.ErrNoRows){return model.CalculationSummary{},ErrNotFound};if err!=nil{return model.CalculationSummary{},err}

 rows,err:=tx.QueryContext(ctx,`SELECT status,COUNT(*) FROM payroll_run_employees WHERE payroll_run_id=? GROUP BY status`,id);if err!=nil{return model.CalculationSummary{},err}
 for rows.Next(){var status string;var n int;if err:=rows.Scan(&status,&n);err!=nil{rows.Close();return model.CalculationSummary{},err};switch status{case "CALCULATED":out.Processed=n;case "FAILED":out.Failed=n;default:out.Pending+=n}}
 if err:=rows.Err();err!=nil{rows.Close();return model.CalculationSummary{},err}
 if err:=rows.Close();err!=nil{return model.CalculationSummary{},err}

 next:="CALCULATED";if out.Failed>0 || out.Pending>0 {next="CALCULATION_FAILED"}
 if _,err=tx.ExecContext(ctx,`UPDATE payroll_runs SET status=? WHERE id=?`,next,id);err!=nil{return model.CalculationSummary{},err}

 if next=="CALCULATED" {
  var userID uint64
  if err:=tx.QueryRowContext(ctx,`SELECT id FROM users WHERE public_id=? AND status='ACTIVE'`,userPublicID).Scan(&userID);err!=nil{
   if errors.Is(err,sql.ErrNoRows){return model.CalculationSummary{},ErrNotFound};return model.CalculationSummary{},err
  }
  if _,err=tx.ExecContext(ctx,`INSERT INTO payroll_run_workflow_events
   (payroll_run_id,actor_user_id,action,from_status,to_status)
   VALUES(?,?,?,?,?)`,id,userID,"CALCULATE",out.Status,next);err!=nil{return model.CalculationSummary{},err}
 }

 if err:=tx.Commit();err!=nil{return model.CalculationSummary{},err}
 out.Status=next
 return out,nil
}
func nilIf(v string) any {if strings.TrimSpace(v)==""{return nil};return strings.TrimSpace(v)}
func mustJSON(v map[string]string) string { b,_:=json.Marshal(v);return string(b) }

func (r Repository) Transition(ctx context.Context, companyPublicID, runPublicID, userPublicID, action string) (model.WorkflowSummary,error) {
 cid,err:=r.companyID(ctx,companyPublicID);if err!=nil{return model.WorkflowSummary{},err}
 tx,err:=r.DB.BeginTx(ctx,nil);if err!=nil{return model.WorkflowSummary{},err};defer tx.Rollback()
 var runID uint64; var out model.WorkflowSummary
 err=tx.QueryRowContext(ctx,`SELECT id,public_id,period_key,period_start,period_end,status,created_at,updated_at
 FROM payroll_runs WHERE company_id=? AND public_id=? FOR UPDATE`,cid,runPublicID).
 Scan(&runID,&out.PublicID,&out.Period,&out.PeriodStart,&out.PeriodEnd,&out.Status,&out.CreatedAt,&out.UpdatedAt)
 if errors.Is(err,sql.ErrNoRows){return model.WorkflowSummary{},ErrNotFound};if err!=nil{return model.WorkflowSummary{},err}
 var userID uint64
 if err:=tx.QueryRowContext(ctx,`SELECT id FROM users WHERE public_id=? AND status='ACTIVE'`,userPublicID).Scan(&userID);err!=nil{
  if errors.Is(err,sql.ErrNoRows){return model.WorkflowSummary{},ErrNotFound};return model.WorkflowSummary{},err
 }
 expected,next:="",""
 switch action {
 case "REVIEW": expected,next="CALCULATED","REVIEW"
 case "APPROVE": expected,next="REVIEW","APPROVED"
 case "FINALIZE": expected,next="APPROVED","FINALIZED"
 case "LOCK": expected,next="FINALIZED","LOCKED"
 default:return model.WorkflowSummary{},ErrConflict
 }
 if out.Status!=expected{return model.WorkflowSummary{},ErrConflict}
 if action=="REVIEW" {
  var total,calculated int
  if err:=tx.QueryRowContext(ctx,`SELECT COUNT(*),COALESCE(SUM(status='CALCULATED'),0) FROM payroll_run_employees WHERE payroll_run_id=?`,runID).Scan(&total,&calculated);err!=nil{return model.WorkflowSummary{},err}
  if total==0 || total!=calculated{return model.WorkflowSummary{},ErrConflict}
 }
 var q string
 switch action {
 case "REVIEW": q=`UPDATE payroll_runs SET status='REVIEW',reviewed_by=?,reviewed_at=UTC_TIMESTAMP() WHERE id=?`
 case "APPROVE": q=`UPDATE payroll_runs SET status='APPROVED',approved_by=?,approved_at=UTC_TIMESTAMP() WHERE id=?`
 case "FINALIZE": q=`UPDATE payroll_runs SET status='FINALIZED',finalized_by=?,finalized_at=UTC_TIMESTAMP() WHERE id=?`
 case "LOCK": q=`UPDATE payroll_runs SET status='LOCKED',locked_by=?,locked_at=UTC_TIMESTAMP() WHERE id=?`
 }
 if _,err:=tx.ExecContext(ctx,q,userID,runID);err!=nil{return model.WorkflowSummary{},err}
 if _,err:=tx.ExecContext(ctx,`INSERT INTO payroll_run_workflow_events
 (payroll_run_id,actor_user_id,action,from_status,to_status)
 VALUES(?,?,?,?,?)`,runID,userID,action,out.Status,next);err!=nil{return model.WorkflowSummary{},err}
 if err:=tx.Commit();err!=nil{return model.WorkflowSummary{},err}
 out.Status=next;out.Action=action;return out,nil
}
