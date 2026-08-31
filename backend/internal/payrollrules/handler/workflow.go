package handler

import (
 "encoding/json";"net/http";"github.com/shopspring/decimal"
 "github.com/alumasinde/budget254-paye-api/internal/middleware"
 rules "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrules/validation"
 "github.com/alumasinde/budget254-paye-api/internal/response"
)

type workflowResponse struct{ Message string `json:"message"` }
type previewRequest struct{ RuleSet rules.RuleSet `json:"rule_set"`; GrossSalary string `json:"gross_salary"` }

func (h Handler) Validate(w http.ResponseWriter,r *http.Request){
 var x rules.RuleSet;if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20)).Decode(&x);err!=nil{response.Fail(w,400,"INVALID_JSON","invalid payload",middleware.ID(r.Context()),nil);return}
 response.JSON(w,200,validation.ValidateRuleSet(x))
}
func (h Handler) Preview(w http.ResponseWriter,r *http.Request){
 var q previewRequest;if err:=json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20)).Decode(&q);err!=nil{response.Fail(w,400,"INVALID_JSON","invalid payload",middleware.ID(r.Context()),nil);return}
 gross,err:=decimal.NewFromString(q.GrossSalary);if err!=nil||gross.IsNegative(){response.Fail(w,422,"INVALID_GROSS_SALARY","invalid gross salary",middleware.ID(r.Context()),nil);return}
 result,err:=servicePreview(q.RuleSet,gross);if err!=nil{response.Fail(w,422,"PREVIEW_FAILED",err.Error(),middleware.ID(r.Context()),nil);return};response.JSON(w,200,result)
}
