package handler

import (
 "encoding/json"; "errors"; "net/http"; "strings"
 "github.com/alumasinde/budget254-paye-api/internal/employee/model"
 repo "github.com/alumasinde/budget254-paye-api/internal/employee/repository"
 "github.com/alumasinde/budget254-paye-api/internal/employee/service"
 companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
 "github.com/alumasinde/budget254-paye-api/internal/middleware"
 "github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct{ Service service.Service }
func (h Handler) Create(w http.ResponseWriter,r *http.Request){var q model.CreateInput;if err:=decode(w,r,&q);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid request");return};v,err:=h.Service.Create(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),q);if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusCreated,v)}
func (h Handler) List(w http.ResponseWriter,r *http.Request){v,err:=h.Service.List(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()));if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,map[string]any{"employees":v})}
func (h Handler) Get(w http.ResponseWriter,r *http.Request){v,err:=h.Service.Get(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("employee_id"));if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,v)}
func (h Handler) Update(w http.ResponseWriter,r *http.Request){var q model.UpdateInput;if err:=decode(w,r,&q);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid request");return};v,err:=h.Service.Update(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("employee_id"),q);if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,v)}
func (h Handler) SalaryHistory(w http.ResponseWriter,r *http.Request){v,err:=h.Service.SalaryHistory(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("employee_id"));if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusOK,map[string]any{"salary_history":v})}
func (h Handler) AddSalary(w http.ResponseWriter,r *http.Request){var q model.SalaryInput;if err:=decode(w,r,&q);err!=nil{fail(w,r,http.StatusBadRequest,"INVALID_REQUEST","invalid request");return};v,err:=h.Service.AddSalary(r.Context(),r.PathValue("company_id"),middleware.UserID(r.Context()),r.PathValue("employee_id"),q);if err!=nil{writeError(w,r,err);return};response.JSON(w,http.StatusCreated,v)}
func decode(w http.ResponseWriter,r *http.Request,v any)error{d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1<<20));d.DisallowUnknownFields();return d.Decode(v)}
func writeError(w http.ResponseWriter,r *http.Request,err error){switch{case errors.Is(err,companyRepo.ErrForbidden):fail(w,r,http.StatusForbidden,"FORBIDDEN","you do not have permission for this action");case errors.Is(err,repo.ErrNotFound):fail(w,r,http.StatusNotFound,"NOT_FOUND","requested resource was not found");case errors.Is(err,repo.ErrConflict):fail(w,r,http.StatusConflict,"CONFLICT","a record with those details already exists");default:m:=strings.TrimSpace(err.Error());if m==""{m="request could not be completed"};fail(w,r,http.StatusUnprocessableEntity,"REQUEST_FAILED",m)}}
func fail(w http.ResponseWriter,r *http.Request,status int,code,message string){response.Fail(w,status,code,message,middleware.ID(r.Context()),nil)}