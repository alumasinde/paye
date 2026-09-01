package repository

import (
 "context"
 "database/sql"
 "errors"
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

func (r Repository) Create(ctx context.Context, companyPublicID, period string, start,end time.Time) (model.Detail,error) {
 cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return model.Detail{},err}
 tx,err:=r.DB.BeginTx(ctx,nil); if err!=nil{return model.Detail{},err}; defer tx.Rollback()
 publicID:=uuid.NewString()
 _,err=tx.ExecContext(ctx,`INSERT INTO payroll_runs(public_id,company_id,period_key,period_start,period_end,status)
 VALUES(?,?,?,?,?,'DRAFT')`,publicID,cid,period,start,end)
 if err!=nil {
  if strings.Contains(strings.ToLower(err.Error()),"duplicate"){return model.Detail{},ErrConflict}
  return model.Detail{},err
 }
 rows,err:=tx.QueryContext(ctx,`SELECT e.id,e.public_id,e.employee_number,e.first_name,COALESCE(e.middle_name,''),e.last_name,
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
 ORDER BY e.first_name,e.last_name`,end,end,cid,end,start)
 if err!=nil{return model.Detail{},err}
 defer rows.Close()
 for rows.Next(){
  var employeeID uint64; var employeePublicID,number,first,middle,last,salary,frequency string
  if err:=rows.Scan(&employeeID,&employeePublicID,&number,&first,&middle,&last,&salary,&frequency);err!=nil{return model.Detail{},err}
  _,err=tx.ExecContext(ctx,`INSERT INTO payroll_run_employees(public_id,payroll_run_id,employee_id,employee_number,first_name,middle_name,last_name,basic_salary,pay_frequency,status)
  VALUES(?,?,?,?,?,?,?,?,?,'PENDING')`,uuid.NewString(),publicID,employeeID,number,first,nilIf(middle),last,salary,frequency)
  if err!=nil{return model.Detail{},err}
 }
 if err:=rows.Err();err!=nil{return model.Detail{},err}
 if err:=tx.Commit();err!=nil{return model.Detail{},err}
 return r.Get(ctx,companyPublicID,publicID)
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
 rows,err:=r.DB.QueryContext(ctx,`SELECT pre.public_id,e.public_id,pre.employee_number,pre.first_name,COALESCE(pre.middle_name,''),pre.last_name,CAST(pre.basic_salary AS CHAR),pre.pay_frequency,pre.status
 FROM payroll_run_employees pre JOIN employees e ON e.id=pre.employee_id WHERE pre.payroll_run_id=? ORDER BY pre.first_name,pre.last_name`,id)
 if err!=nil{return model.Detail{},err};defer rows.Close()
 for rows.Next(){var e model.PayrollRunEmployee;if err:=rows.Scan(&e.PublicID,&e.EmployeeID,&e.EmployeeNumber,&e.FirstName,&e.MiddleName,&e.LastName,&e.BasicSalary,&e.PayFrequency,&e.Status);err!=nil{return model.Detail{},err};d.Employees=append(d.Employees,e)}
 if err:=rows.Err();err!=nil{return model.Detail{},err};d.EmployeeCount=len(d.Employees);return d,nil
}

func nilIf(v string) any {if strings.TrimSpace(v)==""{return nil};return strings.TrimSpace(v)}
