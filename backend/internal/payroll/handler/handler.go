package handler
import (
 "encoding/json"; "net/http"; "strings"; "time"
 "github.com/shopspring/decimal"
 "github.com/alumasinde/budget254-paye-api/internal/middleware"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/dto"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/model"
 "github.com/alumasinde/budget254-paye-api/internal/payroll/service"
 "github.com/alumasinde/budget254-paye-api/internal/response"
)
type Handler struct{ Service service.Service; MaxCustom int; MinDate time.Time }
func (h Handler) Calculate(w http.ResponseWriter,r *http.Request){
 var req dto.CalculateRequest; dec:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20)); dec.DisallowUnknownFields()
 if err:=dec.Decode(&req);err!=nil{response.Fail(w,400,"INVALID_JSON","invalid request body",middleware.ID(r.Context()),nil);return}
 fields:=map[string]string{}; gross,err:=decimal.NewFromString(req.GrossSalary);if err!=nil||!gross.IsPositive(){fields["gross_salary"]="must be a positive decimal amount"}
 date,err:=time.Parse("2006-01-02",req.CalculationDate);if err!=nil{fields["calculation_date"]="must be YYYY-MM-DD"} else if date.Before(h.MinDate){fields["calculation_date"]="date is outside supported history"}
 if len(req.CustomDeductions)>h.MaxCustom{fields["custom_deductions"]="too many custom deductions"}
 custom:=[]model.CustomDeduction{}
 for i,d:=range req.CustomDeductions{key:="custom_deductions["+string(rune('0'+i))+"]";name:=strings.TrimSpace(d.Name);amt,e:=decimal.NewFromString(d.Amount);typ:=model.DeductionType(d.Type)
  if name==""||len(name)>100{fields[key+".name"]="required, max 100 characters"};if e!=nil||amt.IsNegative(){fields[key+".amount"]="must be zero or positive decimal"};if typ!=model.NetPay&&typ!=model.TaxableIncome{fields[key+".type"]="must be NET_PAY or TAXABLE_INCOME"}
  if name!=""&&e==nil&&amt.IsPositive()&&(typ==model.NetPay||typ==model.TaxableIncome){custom=append(custom,model.CustomDeduction{Name:name,Amount:amt,Type:typ})}
 }
 if len(fields)>0{response.Fail(w,422,"VALIDATION_ERROR","request validation failed",middleware.ID(r.Context()),fields);return}
 out,err:=h.Service.Calculate(r.Context(),service.Input{Gross:gross,Date:date,Custom:custom});if err!=nil{response.Fail(w,500,"CALCULATION_ERROR","calculation could not be completed",middleware.ID(r.Context()),nil);return}
 resp:=dto.CalculateResponse{CalculationDate:date.Format("2006-01-02"),GrossSalary:out.Gross.StringFixed(2),TaxableIncome:out.TaxableIncome.StringFixed(2),PAYEBeforeRelief:out.PAYEBeforeRelief.StringFixed(2),Relief:out.Relief.StringFixed(2),PAYE:out.PAYE.StringFixed(2),TotalDeductions:out.TotalDeductions.StringFixed(2),NetSalary:out.Net.StringFixed(2),RuleVersions:out.RuleVersions}
 for _,x:=range out.Statutory{resp.StatutoryDeductions=append(resp.StatutoryDeductions,dto.Deduction{x.Code,x.Name,x.Amount.StringFixed(2),x.ReducesTaxableIncome})}
 for _,x:=range out.Custom{resp.CustomDeductions=append(resp.CustomDeductions,dto.Deduction{x.Code,x.Name,x.Amount.StringFixed(2),x.ReducesTaxableIncome})}
 if req.Explain{for _,t:=range out.Trace{var to *string;if t.To!=nil{s:=t.To.StringFixed(2);to=&s};resp.Trace=append(resp.Trace,dto.BandTrace{t.From.StringFixed(2),to,t.Rate.StringFixed(4),t.Tax.StringFixed(2)})}}
 response.JSON(w,200,resp)
}
