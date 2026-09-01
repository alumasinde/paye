package service

import (
 "context"
 "errors"
 "fmt"
 "strings"
 "time"

 companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
 payrollModel "github.com/alumasinde/budget254-paye-api/internal/payroll/model"
 payrollService "github.com/alumasinde/budget254-paye-api/internal/payroll/service"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrun/model"
 repo "github.com/alumasinde/budget254-paye-api/internal/payrollrun/repository"
 "github.com/shopspring/decimal"
)

type Service struct{ Repo repo.Repository; CompanyRepo companyRepo.Repository; Calculator payrollService.Service }

func (s Service) require(ctx context.Context, companyID,userID,permission string) error {
 ok,err:=s.CompanyRepo.HasPermission(ctx,companyID,userID,permission);if err!=nil{return err}
 if !ok{return companyRepo.ErrForbidden};return nil
}

func (s Service) Create(ctx context.Context,companyID,userID string,in model.CreateInput)(model.Detail,error){
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.Detail{},err}
 period:=strings.TrimSpace(in.Period)
 start,end,err:=parsePeriod(period);if err!=nil{return model.Detail{},err}
 return s.Repo.Create(ctx,companyID,period,start,end)
}

func (s Service) List(ctx context.Context,companyID,userID string)([]model.PayrollRun,error){
 if err:=s.require(ctx,companyID,userID,"employees.read");err!=nil{return nil,err}
 return s.Repo.List(ctx,companyID)
}

func (s Service) Get(ctx context.Context,companyID,userID,runID string)(model.Detail,error){
 if err:=s.require(ctx,companyID,userID,"employees.read");err!=nil{return model.Detail{},err}
 return s.Repo.Get(ctx,companyID,runID)
}

func (s Service) Calculate(ctx context.Context,companyID,userID,runID string)(model.CalculationSummary,error){
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.CalculationSummary{},err}
 run,employees,err:=s.Repo.Pending(ctx,companyID,runID);if err!=nil{return model.CalculationSummary{},err}
 calculationDate:=run.PeriodEnd
 for _,employee:=range employees {
  gross,parseErr:=decimal.NewFromString(employee.BasicSalary)
  if parseErr!=nil || !gross.IsPositive() {
   _=s.Repo.SaveFailure(ctx,companyID,runID,employee.PublicID,"employee has an invalid basic salary snapshot")
   continue
  }
  customInputs:=make([]payrollModel.CustomDeduction,0,len(employee.Adjustments))
  customNet:=decimal.Zero
  taxableDeductions:=decimal.Zero
  nonTaxableEarnings:=decimal.Zero
  invalidAdjustment:=false
  for _,adjustment:=range employee.Adjustments {
   amount,amountErr:=decimal.NewFromString(adjustment.Amount)
   if amountErr!=nil || !amount.IsPositive() {
    _=s.Repo.SaveFailure(ctx,companyID,runID,employee.PublicID,"employee has an invalid payroll adjustment")
    invalidAdjustment=true
    break
   }
   switch adjustment.Kind {
   case "EARNING":
    gross=gross.Add(amount)
    if !adjustment.Taxable {
     nonTaxableEarnings=nonTaxableEarnings.Add(amount)
     customInputs=append(customInputs,payrollModel.CustomDeduction{Name:adjustment.Name,Amount:amount,Type:payrollModel.TaxableIncome})
    }
   case "DEDUCTION":
    deductionType:=payrollModel.NetPay
    if adjustment.ReducesTaxableIncome { deductionType=payrollModel.TaxableIncome; taxableDeductions=taxableDeductions.Add(amount) } else { customNet=customNet.Add(amount) }
    customInputs=append(customInputs,payrollModel.CustomDeduction{Name:adjustment.Name,Amount:amount,Type:deductionType})
   default:
    _=s.Repo.SaveFailure(ctx,companyID,runID,employee.PublicID,"employee has an unsupported payroll adjustment")
    invalidAdjustment=true
    break
   }
  }
  if invalidAdjustment { continue }
  out,calcErr:=s.Calculator.Calculate(ctx,payrollService.Input{Gross:gross,Date:calculationDate,Custom:customInputs})
  if calcErr!=nil {
   _=s.Repo.SaveFailure(ctx,companyID,runID,employee.PublicID,safeError(calcErr))
   continue
  }
  if !nonTaxableEarnings.IsZero() {
   out.TotalDeductions=out.TotalDeductions.Sub(nonTaxableEarnings)
   out.Net=out.Net.Add(nonTaxableEarnings)
  }
  statutory:=decimal.Zero
  for _,d:=range out.Statutory { statutory=statutory.Add(d.Amount) }
  custom:=customNet.Add(taxableDeductions)
  employee.GrossSalary=out.Gross.StringFixed(2)
  employee.TaxableIncome=out.TaxableIncome.StringFixed(2)
  employee.PAYEBeforeRelief=out.PAYEBeforeRelief.StringFixed(2)
  employee.Relief=out.Relief.StringFixed(2)
  employee.PAYE=out.PAYE.StringFixed(2)
  employee.StatutoryDeductions=statutory.StringFixed(2)
  employee.CustomDeductions=custom.StringFixed(2)
  employee.TotalDeductions=out.TotalDeductions.StringFixed(2)
  employee.NetSalary=out.Net.StringFixed(2)
  employee.RuleVersions=out.RuleVersions
  if err:=s.Repo.SaveCalculation(ctx,companyID,runID,employee.PublicID,employee);err!=nil{return model.CalculationSummary{},err}
 }
 return s.Repo.FinalizeCalculation(ctx,companyID,runID,userID)
}

func (s Service) AddAdjustment(ctx context.Context,companyID,userID,runID,employeeRunID string,in model.AdjustmentInput)(model.PayrollAdjustment,error){
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.PayrollAdjustment{},err}
 in,err:=validateAdjustment(in);if err!=nil{return model.PayrollAdjustment{},err}
 return s.Repo.AddAdjustment(ctx,companyID,runID,employeeRunID,in)
}
func (s Service) UpdateAdjustment(ctx context.Context,companyID,userID,runID,employeeRunID,adjustmentID string,in model.AdjustmentInput)(model.PayrollAdjustment,error){
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.PayrollAdjustment{},err}
 in,err:=validateAdjustment(in);if err!=nil{return model.PayrollAdjustment{},err}
 return s.Repo.UpdateAdjustment(ctx,companyID,runID,employeeRunID,adjustmentID,in)
}
func (s Service) DeleteAdjustment(ctx context.Context,companyID,userID,runID,employeeRunID,adjustmentID string) error{
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return err}
 return s.Repo.DeleteAdjustment(ctx,companyID,runID,employeeRunID,adjustmentID)
}
func (s Service) AddBulkInput(ctx context.Context,companyID,userID,runID string,in model.BulkInput)(model.BulkInputResult,error){if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.BulkInputResult{},err};a,err:=validateAdjustment(model.AdjustmentInput{Name:in.Name,Kind:in.Kind,Category:in.Category,Payee:in.Payee,ReferenceNo:in.ReferenceNo,Amount:in.Amount,Taxable:in.Taxable,ReducesTaxableIncome:in.ReducesTaxableIncome});if err!=nil{return model.BulkInputResult{},err};in.Name=a.Name;in.Kind=a.Kind;in.Category=a.Category;in.Payee=a.Payee;in.ReferenceNo=a.ReferenceNo;in.Amount=a.Amount;in.Taxable=a.Taxable;in.ReducesTaxableIncome=a.ReducesTaxableIncome;return s.Repo.AddBulkInput(ctx,companyID,runID,in)}

func validateAdjustment(in model.AdjustmentInput)(model.AdjustmentInput,error){
 in.Name=strings.TrimSpace(in.Name)
 in.Kind=strings.ToUpper(strings.TrimSpace(in.Kind))
 in.Category=strings.ToUpper(strings.TrimSpace(in.Category)); if in.Category==""{in.Category="OTHER"}
 in.Payee=strings.TrimSpace(in.Payee); in.ReferenceNo=strings.TrimSpace(in.ReferenceNo)
 if in.Name=="" || len(in.Name)>150{return in,errors.New("adjustment name is required")}
 if in.Kind!="EARNING" && in.Kind!="DEDUCTION"{return in,errors.New("adjustment kind must be EARNING or DEDUCTION")}
 switch in.Category{case "ALLOWANCE","BONUS","OVERTIME","COMMISSION","WELFARE","SACCO","LOAN","ADVANCE","INSURANCE","OTHER":default:return in,errors.New("unsupported payroll input category")}
 amount,err:=decimal.NewFromString(strings.TrimSpace(in.Amount));if err!=nil || !amount.IsPositive(){return in,errors.New("adjustment amount must be greater than zero")}
 in.Amount=amount.StringFixed(2)
 if in.Kind=="EARNING" {in.ReducesTaxableIncome=false}
 return in,nil
}

func safeError(err error) string { v:=strings.TrimSpace(err.Error());if v==""{return "calculation failed"};if len(v)>500{return v[:500]};return v }

func parsePeriod(v string)(time.Time,time.Time,error){
 if len(v)!=7{return time.Time{},time.Time{},errors.New("period must be YYYY-MM")}
 start,err:=time.Parse("2006-01",v);if err!=nil{return time.Time{},time.Time{},errors.New("period must be YYYY-MM")}
 end:=start.AddDate(0,1,0).Add(-24*time.Hour)
 if start.IsZero() || end.Before(start){return time.Time{},time.Time{},fmt.Errorf("invalid period")}
 return start,end,nil
}

func (s Service) Review(ctx context.Context,companyID,userID,runID string)(model.WorkflowSummary,error){
 if err:=s.require(ctx,companyID,userID,"payroll.review");err!=nil{return model.WorkflowSummary{},err}
 return s.Repo.Transition(ctx,companyID,runID,userID,"REVIEW")
}
func (s Service) Approve(ctx context.Context,companyID,userID,runID string)(model.WorkflowSummary,error){
 if err:=s.require(ctx,companyID,userID,"payroll.approve");err!=nil{return model.WorkflowSummary{},err}
 return s.Repo.Transition(ctx,companyID,runID,userID,"APPROVE")
}
func (s Service) Finalize(ctx context.Context,companyID,userID,runID string)(model.WorkflowSummary,error){
 if err:=s.require(ctx,companyID,userID,"payroll.finalize");err!=nil{return model.WorkflowSummary{},err}
 return s.Repo.Transition(ctx,companyID,runID,userID,"FINALIZE")
}
func (s Service) Lock(ctx context.Context,companyID,userID,runID string)(model.WorkflowSummary,error){
 if err:=s.require(ctx,companyID,userID,"payroll.lock");err!=nil{return model.WorkflowSummary{},err}
 return s.Repo.Transition(ctx,companyID,runID,userID,"LOCK")
}

func (s Service) RefreshEmployees(ctx context.Context,companyID,userID,runID string)(model.Detail,error){
 if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.Detail{},err}
 return s.Repo.RefreshEmployees(ctx,companyID,runID)
}

func (s Service) Reopen(ctx context.Context,companyID,userID,runID string)(model.WorkflowSummary,error){if err:=s.require(ctx,companyID,userID,"employees.write");err!=nil{return model.WorkflowSummary{},err};return s.Repo.Reopen(ctx,companyID,runID,userID)}
func (s Service) Validate(ctx context.Context,companyID,userID,runID string)(model.ValidationSummary,error){if err:=s.require(ctx,companyID,userID,"employees.read");err!=nil{return model.ValidationSummary{},err};return s.Repo.Validate(ctx,companyID,runID)}
