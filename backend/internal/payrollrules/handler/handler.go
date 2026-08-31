package handler

import (
    "encoding/json"
    "log"
    "net/http"
    "github.com/alumasinde/budget254-paye-api/internal/admin/audit"
    "github.com/alumasinde/budget254-paye-api/internal/middleware"
    "github.com/alumasinde/budget254-paye-api/internal/payrollrules/model"
    repo "github.com/alumasinde/budget254-paye-api/internal/payrollrules/repository"
    "github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct {
    Repo  repo.Repository
    Audit audit.Writer
}

func (h Handler) List(w http.ResponseWriter,r *http.Request) {
    x,err:=h.Repo.List(r.Context());if err!=nil{response.JSON(w,500,map[string]string{"code":"RULES_LIST_FAILED","message":"Could not load rules"});return}
    response.JSON(w,200,map[string]any{"items":x})
}
func (h Handler) Create(w http.ResponseWriter,r *http.Request) {
    var x model.RuleSet
    d:=json.NewDecoder(http.MaxBytesReader(w,r.Body,2<<20));d.DisallowUnknownFields()
    if err:=d.Decode(&x);err!=nil{response.JSON(w,400,map[string]string{"code":"INVALID_RULE_SET","message":"Invalid rule set payload"});return}
    adminID:=middleware.AdminDBID(r.Context())
    x,err:=h.Repo.Create(r.Context(),x,adminID)
    if err!=nil{response.JSON(w,422,map[string]string{"code":"RULE_SET_CREATE_FAILED","message":"Could not create rule set"});return}
    if auditErr:=h.Audit.Write(r.Context(),&adminID,"rule_set.create","payroll_rule_set",x.PublicID,middleware.ID(r.Context()),r.RemoteAddr,nil,x);auditErr!=nil{
        log.Printf("audit write failed (rule_set.create %s): %v",x.PublicID,auditErr)
    }
    response.JSON(w,201,x)
}
