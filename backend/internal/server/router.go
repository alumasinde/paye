package server

import (
	"github.com/alumasinde/budget254-paye-api/internal/health"
	payroll "github.com/alumasinde/budget254-paye-api/internal/payroll/routes"
	payrollHandler "github.com/alumasinde/budget254-paye-api/internal/payroll/handler"
	"github.com/alumasinde/budget254-paye-api/internal/periods"
	periodHandler "github.com/alumasinde/budget254-paye-api/internal/periods/handler"
	"github.com/alumasinde/budget254-paye-api/internal/rules"
	ruleHandler "github.com/alumasinde/budget254-paye-api/internal/rules/handler"
	"net/http"
)

func NewRouter(healthHandler health.Handler, rh *ruleHandler.Handler, ph payrollHandler.Handler) *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("GET /health", healthHandler.Liveness)
	m.HandleFunc("GET /api/v1/health", healthHandler.Liveness)
	m.HandleFunc("GET /api/v1/ready", healthHandler.Readiness)
	periods.RegisterRoutes(m, periodHandler.Handler{})
	rules.RegisterRoutes(m, rh)
	payroll.Register(m, ph)
	return m
}
