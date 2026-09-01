package handler

import (
 "encoding/json"
 "errors"
 "net/http"
 "strings"

 companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
 "github.com/alumasinde/budget254-paye-api/internal/middleware"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrun/model"
 repo "github.com/alumasinde/budget254-paye-api/internal/payrollrun/repository"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrun/service"
 "github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct{ Service service.Service }

func (h Handler) Create(w http.ResponseWriter,r *http.Request){
 var in model.CreateInput
 d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20));d.DisallowUnknownFields()
 if err:=d.Decode(&in);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid request");return}
 out,err:=h.Service.Create(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),in)
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusCreated,out)
}
func (h Handler) List(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.List(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,map[string]any{"payroll_runs":out})
}
func (h Handler) Get(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Get(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}
func (h Handler) Calculate(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Calculate(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return}
 status:=http.StatusOK;if out.Failed>0 || out.Pending>0 {status=http.StatusUnprocessableEntity}
 response.JSON(w,status,out)
}
func (h Handler) AddAdjustment(w http.ResponseWriter,r *http.Request){
 var in model.AdjustmentInput
 d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20));d.DisallowUnknownFields()
 if err:=d.Decode(&in);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid adjustment request");return}
 out,err:=h.Service.AddAdjustment(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"),r.PathValue("employee_run_id"),in)
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusCreated,out)
}
func (h Handler) UpdateAdjustment(w http.ResponseWriter,r *http.Request){
 var in model.AdjustmentInput
 d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20));d.DisallowUnknownFields()
 if err:=d.Decode(&in);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid adjustment request");return}
 out,err:=h.Service.UpdateAdjustment(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"),r.PathValue("employee_run_id"),r.PathValue("adjustment_id"),in)
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}
func (h Handler) DeleteAdjustment(w http.ResponseWriter,r *http.Request){
 err:=h.Service.DeleteAdjustment(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"),r.PathValue("employee_run_id"),r.PathValue("adjustment_id"))
 if err!=nil{writeError(w,r,err);return};w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter,r *http.Request,err error){
 switch{
 case errors.Is(err,companyRepo.ErrForbidden):fail(w,r,http.StatusForbidden,"FORBIDDEN","you do not have permission for this action")
 case errors.Is(err,repo.ErrNotFound):fail(w,r,http.StatusNotFound,"NOT_FOUND","requested resource was not found")
 case errors.Is(err,repo.ErrConflict):fail(w,r,http.StatusConflict,"CONFLICT","payroll run cannot be calculated in its current state")
 default:m:=strings.TrimSpace(err.Error());if m==""{m="request could not be completed"};fail(w,r,http.StatusUnprocessableEntity,"REQUEST_FAILED",m)
 }
}
func fail(w http.ResponseWriter,r *http.Request,status int,code,message string){response.Fail(w,status,code,message,middleware.ID(r.Context()),nil)}

func (h Handler) Review(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Review(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}
func (h Handler) Approve(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Approve(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}
func (h Handler) Finalize(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Finalize(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}
func (h Handler) Lock(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.Lock(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}

func (h Handler) RefreshEmployees(w http.ResponseWriter,r *http.Request){
 out,err:=h.Service.RefreshEmployees(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"))
 if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)
}

func (h Handler) Reopen(w http.ResponseWriter,r *http.Request){out,err:=h.Service.Reopen(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"));if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)}
func (h Handler) Validate(w http.ResponseWriter,r *http.Request){out,err:=h.Service.Validate(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("payroll_run_id"));if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,out)}
