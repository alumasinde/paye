package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/alumasinde/budget254-paye-api/internal/employee/model"
)

var ( ErrNotFound=errors.New("not found"); ErrConflict=errors.New("conflict") )

type Repository struct{ DB *sql.DB }

func (r Repository) companyID(ctx context.Context, publicID string)(uint64,error){ var id uint64; err:=r.DB.QueryRowContext(ctx,"SELECT id FROM companies WHERE public_id=? AND status='ACTIVE'",publicID).Scan(&id); if errors.Is(err,sql.ErrNoRows){return 0,ErrNotFound}; return id,err }

func (r Repository) departmentID(ctx context.Context, companyID uint64, publicID string)(any,error){ if strings.TrimSpace(publicID)==""{return nil,nil}; var id uint64; err:=r.DB.QueryRowContext(ctx,"SELECT id FROM departments WHERE public_id=? AND company_id=? AND is_active=TRUE",publicID,companyID).Scan(&id); if errors.Is(err,sql.ErrNoRows){return nil,ErrNotFound}; return id,err }

func (r Repository) Create(ctx context.Context, companyPublicID string, in model.CreateInput)(model.EmployeeDetail,error){
	companyID,err:=r.companyID(ctx,companyPublicID); if err!=nil{return model.EmployeeDetail{},err}
	departmentID,err:=r.departmentID(ctx,companyID,in.DepartmentID); if err!=nil{return model.EmployeeDetail{},err}
	tx,err:=r.DB.BeginTx(ctx,nil); if err!=nil{return model.EmployeeDetail{},err}; defer tx.Rollback()
	pid:=uuid.NewString()
	_,err=tx.ExecContext(ctx,`INSERT INTO employees(public_id,company_id,employee_number,first_name,middle_name,last_name,gender,date_of_birth,national_id,passport_number,nationality,kra_pin,nssf_number,shif_number,nhif_number,employment_date,job_title,department_id,employment_type)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,pid,companyID,in.EmployeeNumber,in.FirstName,nilIf(in.MiddleName),in.LastName,nilIf(in.Gender),dateOrNil(in.DateOfBirth),nilIf(in.NationalID),nilIf(in.PassportNumber),nilIf(in.Nationality),nilIf(in.KRAPIN),nilIf(in.NSSFNumber),nilIf(in.SHIFNumber),nilIf(in.NHIFNumber),in.EmploymentDate,nilIf(in.JobTitle),departmentID,nilIf(in.EmploymentType))
	if err!=nil { if strings.Contains(strings.ToLower(err.Error()),"duplicate"){return model.EmployeeDetail{},ErrConflict}; return model.EmployeeDetail{},err }
	var employeeID uint64; if err=tx.QueryRowContext(ctx,"SELECT id FROM employees WHERE public_id=?",pid).Scan(&employeeID); err!=nil{return model.EmployeeDetail{},err}
	if _,err=tx.ExecContext(ctx,`INSERT INTO employee_salary_history(public_id,employee_id,basic_salary,pay_frequency,effective_from) VALUES(?,?,?,?,?)`,uuid.NewString(),employeeID,in.BasicSalary,in.PayFrequency,in.EffectiveFrom); err!=nil{return model.EmployeeDetail{},err}
	if err=tx.Commit(); err!=nil{return model.EmployeeDetail{},err}
	return r.Get(ctx,companyPublicID,pid)
}

func (r Repository) List(ctx context.Context, companyPublicID string)([]model.Employee,error){ cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return nil,err}; rows,err:=r.DB.QueryContext(ctx,`SELECT public_id,employee_number,first_name,COALESCE(middle_name,''),last_name,COALESCE(gender,''),date_of_birth,COALESCE(national_id,''),COALESCE(passport_number,''),COALESCE(nationality,''),COALESCE(kra_pin,''),COALESCE(nssf_number,''),COALESCE(shif_number,''),COALESCE(nhif_number,''),employment_date,termination_date,COALESCE(job_title,''),COALESCE((SELECT d.name FROM departments d WHERE d.id=employees.department_id),''),COALESCE(employment_type,''),status,created_at FROM employees WHERE company_id=? ORDER BY first_name,last_name`,cid); if err!=nil{return nil,err}; defer rows.Close(); out:=[]model.Employee{}; for rows.Next(){var e model.Employee; if err:=rows.Scan(&e.PublicID,&e.EmployeeNumber,&e.FirstName,&e.MiddleName,&e.LastName,&e.Gender,&e.DateOfBirth,&e.NationalID,&e.PassportNumber,&e.Nationality,&e.KRAPIN,&e.NSSFNumber,&e.SHIFNumber,&e.NHIFNumber,&e.EmploymentDate,&e.TerminationDate,&e.JobTitle,&e.Department,&e.EmploymentType,&e.Status,&e.CreatedAt);err!=nil{return nil,err}; out=append(out,e)}; return out,rows.Err() }

func (r Repository) Get(ctx context.Context, companyPublicID, employeePublicID string)(model.EmployeeDetail,error){ cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return model.EmployeeDetail{},err}; var e model.Employee; err=r.DB.QueryRowContext(ctx,`SELECT public_id,employee_number,first_name,COALESCE(middle_name,''),last_name,COALESCE(gender,''),date_of_birth,COALESCE(national_id,''),COALESCE(passport_number,''),COALESCE(nationality,''),COALESCE(kra_pin,''),COALESCE(nssf_number,''),COALESCE(shif_number,''),COALESCE(nhif_number,''),employment_date,termination_date,COALESCE(job_title,''),COALESCE((SELECT d.name FROM departments d WHERE d.id=employees.department_id),''),COALESCE(employment_type,''),status,created_at FROM employees WHERE company_id=? AND public_id=?`,cid,employeePublicID).Scan(&e.PublicID,&e.EmployeeNumber,&e.FirstName,&e.MiddleName,&e.LastName,&e.Gender,&e.DateOfBirth,&e.NationalID,&e.PassportNumber,&e.Nationality,&e.KRAPIN,&e.NSSFNumber,&e.SHIFNumber,&e.NHIFNumber,&e.EmploymentDate,&e.TerminationDate,&e.JobTitle,&e.Department,&e.EmploymentType,&e.Status,&e.CreatedAt); if errors.Is(err,sql.ErrNoRows){return model.EmployeeDetail{},ErrNotFound}; if err!=nil{return model.EmployeeDetail{},err}; d:=model.EmployeeDetail{Employee:e}; var s model.Salary; err=r.DB.QueryRowContext(ctx,`SELECT public_id,CAST(basic_salary AS CHAR),pay_frequency,effective_from,effective_to FROM employee_salary_history WHERE employee_id=(SELECT id FROM employees WHERE public_id=?) AND effective_from<=CURDATE() AND (effective_to IS NULL OR effective_to>=CURDATE()) ORDER BY effective_from DESC LIMIT 1`,employeePublicID).Scan(&s.PublicID,&s.BasicSalary,&s.PayFrequency,&s.EffectiveFrom,&s.EffectiveTo); if err==nil{d.CurrentSalary=&s}else if !errors.Is(err,sql.ErrNoRows){return model.EmployeeDetail{},err}; return d,nil }

func (r Repository) Update(ctx context.Context, companyPublicID, employeePublicID string, in model.UpdateInput)(model.EmployeeDetail,error){ cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return model.EmployeeDetail{},err}; departmentID,err:=r.departmentID(ctx,cid,in.DepartmentID); if err!=nil{return model.EmployeeDetail{},err}; res,err:=r.DB.ExecContext(ctx,`UPDATE employees SET first_name=?,middle_name=?,last_name=?,gender=?,date_of_birth=?,national_id=?,passport_number=?,nationality=?,kra_pin=?,nssf_number=?,shif_number=?,nhif_number=?,employment_date=?,termination_date=?,job_title=?,department_id=?,employment_type=?,status=? WHERE company_id=? AND public_id=?`,in.FirstName,nilIf(in.MiddleName),in.LastName,nilIf(in.Gender),dateOrNil(in.DateOfBirth),nilIf(in.NationalID),nilIf(in.PassportNumber),nilIf(in.Nationality),nilIf(in.KRAPIN),nilIf(in.NSSFNumber),nilIf(in.SHIFNumber),nilIf(in.NHIFNumber),in.EmploymentDate,dateOrNil(in.TerminationDate),nilIf(in.JobTitle),departmentID,nilIf(in.EmploymentType),in.Status,cid,employeePublicID); if err!=nil{return model.EmployeeDetail{},err}; n,_:=res.RowsAffected(); if n==0{return model.EmployeeDetail{},ErrNotFound}; return r.Get(ctx,companyPublicID,employeePublicID) }

func (r Repository) SalaryHistory(ctx context.Context, companyPublicID, employeePublicID string)([]model.Salary,error){
	cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return nil,err}
	rows,err:=r.DB.QueryContext(ctx,`SELECT public_id,CAST(basic_salary AS CHAR),pay_frequency,effective_from,effective_to FROM employee_salary_history WHERE employee_id=(SELECT id FROM employees WHERE company_id=? AND public_id=?) ORDER BY effective_from DESC`,cid,employeePublicID)
	if err!=nil{return nil,err}; defer rows.Close()
	out:=[]model.Salary{}
	for rows.Next(){var s model.Salary;if err:=rows.Scan(&s.PublicID,&s.BasicSalary,&s.PayFrequency,&s.EffectiveFrom,&s.EffectiveTo);err!=nil{return nil,err};out=append(out,s)}
	if err:=rows.Err();err!=nil{return nil,err}
	if len(out)==0 { var exists int; err=r.DB.QueryRowContext(ctx,"SELECT 1 FROM employees WHERE company_id=? AND public_id=?",cid,employeePublicID).Scan(&exists); if errors.Is(err,sql.ErrNoRows){return nil,ErrNotFound}; if err!=nil{return nil,err} }
	return out,nil
}

func (r Repository) AddSalary(ctx context.Context, companyPublicID, employeePublicID string, in model.SalaryInput)(model.Salary,error){
	cid,err:=r.companyID(ctx,companyPublicID); if err!=nil{return model.Salary{},err}
	tx,err:=r.DB.BeginTx(ctx,nil); if err!=nil{return model.Salary{},err}; defer tx.Rollback()
	var employeeID uint64
	err=tx.QueryRowContext(ctx,"SELECT id FROM employees WHERE company_id=? AND public_id=? FOR UPDATE",cid,employeePublicID).Scan(&employeeID)
	if errors.Is(err,sql.ErrNoRows){return model.Salary{},ErrNotFound}; if err!=nil{return model.Salary{},err}
	var latest time.Time
	err=tx.QueryRowContext(ctx,"SELECT effective_from FROM employee_salary_history WHERE employee_id=? ORDER BY effective_from DESC LIMIT 1 FOR UPDATE",employeeID).Scan(&latest)
	if err!=nil{return model.Salary{},err}
	next,err:=time.Parse("2006-01-02",in.EffectiveFrom); if err!=nil{return model.Salary{},err}
	if !next.After(latest){return model.Salary{},errors.New("salary effective_from must be after the latest salary effective date")}
	previousEnd:=next.AddDate(0,0,-1).Format("2006-01-02")
	if _,err=tx.ExecContext(ctx,"UPDATE employee_salary_history SET effective_to=? WHERE employee_id=? AND effective_from=? AND effective_to IS NULL",previousEnd,employeeID,latest.Format("2006-01-02"));err!=nil{return model.Salary{},err}
	pid:=uuid.NewString()
	if _,err=tx.ExecContext(ctx,"INSERT INTO employee_salary_history(public_id,employee_id,basic_salary,pay_frequency,effective_from) VALUES(?,?,?,?,?)",pid,employeeID,in.BasicSalary,in.PayFrequency,in.EffectiveFrom);err!=nil{return model.Salary{},err}
	if err=tx.Commit();err!=nil{return model.Salary{},err}
	return model.Salary{PublicID:pid,BasicSalary:in.BasicSalary,PayFrequency:in.PayFrequency,EffectiveFrom:next},nil
}

func nilIf(s string) any { if strings.TrimSpace(s)=="" { return nil }; return strings.TrimSpace(s) }
func dateOrNil(s string) any { if strings.TrimSpace(s)=="" { return nil }; return s }
func _time() time.Time { return time.Time{} }