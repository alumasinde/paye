package service

import (
 "context"; "errors"; "strconv"; "strings"; "time"
 "github.com/alumasinde/budget254-paye-api/internal/employee/model"
 repo "github.com/alumasinde/budget254-paye-api/internal/employee/repository"
 companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
)

type Service struct { Repo repo.Repository; CompanyRepo companyRepo.Repository }

func (s Service) require(ctx context.Context,c,u,p string) error { ok,err:=s.CompanyRepo.HasPermission(ctx,c,u,p); if err!=nil{return err}; if !ok{return companyRepo.ErrForbidden}; return nil }
func (s Service) Create(ctx context.Context,c,u string,in model.CreateInput)(model.EmployeeDetail,error){ if err:=s.require(ctx,c,u,"employees.write");err!=nil{return model.EmployeeDetail{},err}; normalizeCreate(&in); if err:=validateCreate(in);err!=nil{return model.EmployeeDetail{},err}; return s.Repo.Create(ctx,c,in) }
func (s Service) List(ctx context.Context,c,u string)([]model.Employee,error){ if err:=s.require(ctx,c,u,"employees.read");err!=nil{return nil,err}; return s.Repo.List(ctx,c) }
func (s Service) Get(ctx context.Context,c,u,e string)(model.EmployeeDetail,error){ if err:=s.require(ctx,c,u,"employees.read");err!=nil{return model.EmployeeDetail{},err}; return s.Repo.Get(ctx,c,e) }
func (s Service) Update(ctx context.Context,c,u,e string,in model.UpdateInput)(model.EmployeeDetail,error){ if err:=s.require(ctx,c,u,"employees.write");err!=nil{return model.EmployeeDetail{},err}; normalizeUpdate(&in); if err:=validateUpdate(in);err!=nil{return model.EmployeeDetail{},err}; return s.Repo.Update(ctx,c,e,in) }

func clean(v string) string{return strings.TrimSpace(v)}
func normalizeCreate(in *model.CreateInput){ in.EmployeeNumber=clean(in.EmployeeNumber);in.FirstName=clean(in.FirstName);in.MiddleName=clean(in.MiddleName);in.LastName=clean(in.LastName);in.Gender=strings.ToUpper(clean(in.Gender));in.EmploymentDate=clean(in.EmploymentDate);in.BasicSalary=clean(in.BasicSalary);in.PayFrequency=strings.ToUpper(clean(in.PayFrequency));in.EffectiveFrom=clean(in.EffectiveFrom);in.KRAPIN=strings.ToUpper(clean(in.KRAPIN)) }
func normalizeUpdate(in *model.UpdateInput){ in.FirstName=clean(in.FirstName);in.MiddleName=clean(in.MiddleName);in.LastName=clean(in.LastName);in.Gender=strings.ToUpper(clean(in.Gender));in.Status=strings.ToUpper(clean(in.Status));in.EmploymentDate=clean(in.EmploymentDate);in.TerminationDate=clean(in.TerminationDate);in.KRAPIN=strings.ToUpper(clean(in.KRAPIN)) }
func validDate(v string)bool{_,e:=time.Parse("2006-01-02",v);return e==nil}
func validateCreate(in model.CreateInput)error{ if in.EmployeeNumber==""||in.FirstName==""||in.LastName==""||!validDate(in.EmploymentDate)||!validDate(in.EffectiveFrom){return errors.New("employee_number, first_name, last_name, employment_date and effective_from are required")}; n,e:=strconv.ParseFloat(in.BasicSalary,64);if e!=nil||n<0{return errors.New("basic_salary must be a non-negative number")};if in.PayFrequency==""{return errors.New("pay_frequency is required")};return validateGender(in.Gender) }
func validateUpdate(in model.UpdateInput)error{if in.FirstName==""||in.LastName==""||!validDate(in.EmploymentDate){return errors.New("first_name, last_name and employment_date are required")};if in.DateOfBirth!=""&&!validDate(in.DateOfBirth){return errors.New("date_of_birth must be YYYY-MM-DD")};if in.TerminationDate!=""&&!validDate(in.TerminationDate){return errors.New("termination_date must be YYYY-MM-DD")};if in.Status==""{return errors.New("status is required")};return validateGender(in.Gender)}
func validateGender(v string)error{if v==""{return nil};switch v{case "MALE","FEMALE","OTHER","PREFER_NOT_TO_SAY":return nil};return errors.New("invalid gender")}