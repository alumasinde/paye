package service

import (
 "context"
 "errors"
 "strings"
 "time"

 companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
 "github.com/alumasinde/budget254-paye-api/internal/payrollrun/model"
 repo "github.com/alumasinde/budget254-paye-api/internal/payrollrun/repository"
)

type Service struct{ Repo repo.Repository; CompanyRepo companyRepo.Repository }

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

func parsePeriod(v string)(time.Time,time.Time,error){
 if len(v)!=7{return time.Time{},time.Time{},errors.New("period must be YYYY-MM")}
 start,err:=time.Parse("2006-01",v);if err!=nil{return time.Time{},time.Time{},errors.New("period must be YYYY-MM")}
 end:=start.AddDate(0,1,0).Add(-24*time.Hour)
 return start,end,nil
}
