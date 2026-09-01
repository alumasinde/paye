package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alumasinde/budget254-paye-api/internal/admin/audit"
	adminHandler "github.com/alumasinde/budget254-paye-api/internal/admin/handler"
	adminRepo "github.com/alumasinde/budget254-paye-api/internal/admin/repository"
	adminService "github.com/alumasinde/budget254-paye-api/internal/admin/service"
	authHandler "github.com/alumasinde/budget254-paye-api/internal/auth/handler"
	authRepo "github.com/alumasinde/budget254-paye-api/internal/auth/repository"
	authService "github.com/alumasinde/budget254-paye-api/internal/auth/service"
	"github.com/alumasinde/budget254-paye-api/internal/config"
	companyHandler "github.com/alumasinde/budget254-paye-api/internal/company/handler"
	companyRepo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
	companyService "github.com/alumasinde/budget254-paye-api/internal/company/service"
	departmentHandler "github.com/alumasinde/budget254-paye-api/internal/department/handler"
	departmentRepo "github.com/alumasinde/budget254-paye-api/internal/department/repository"
	departmentService "github.com/alumasinde/budget254-paye-api/internal/department/service"
	
	employeeHandler "github.com/alumasinde/budget254-paye-api/internal/employee/handler"
	employeeRepo "github.com/alumasinde/budget254-paye-api/internal/employee/repository"
	employeeService "github.com/alumasinde/budget254-paye-api/internal/employee/service"
	payrollRunHandler "github.com/alumasinde/budget254-paye-api/internal/payrollrun/handler"
	payrollRunRepo "github.com/alumasinde/budget254-paye-api/internal/payrollrun/repository"
	payrollRunService "github.com/alumasinde/budget254-paye-api/internal/payrollrun/service"
	"github.com/alumasinde/budget254-paye-api/internal/database"
	"github.com/alumasinde/budget254-paye-api/internal/health"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	ph "github.com/alumasinde/budget254-paye-api/internal/payroll/handler"
	ps "github.com/alumasinde/budget254-paye-api/internal/payroll/service"
	rulesHandler "github.com/alumasinde/budget254-paye-api/internal/payrollrules/handler"
	rulesRepo "github.com/alumasinde/budget254-paye-api/internal/payrollrules/repository"
	rulesWorkflow "github.com/alumasinde/budget254-paye-api/internal/payrollrules/service"
	rh "github.com/alumasinde/budget254-paye-api/internal/rules/handler"
	rr "github.com/alumasinde/budget254-paye-api/internal/rules/repository"
	savedHandler "github.com/alumasinde/budget254-paye-api/internal/saved/handler"
	savedRepo "github.com/alumasinde/budget254-paye-api/internal/saved/repository"
)

type App struct {
	Server          *http.Server
	DB              *database.MySQL
	Log             *slog.Logger
	shutdownTimeout time.Duration
}

func NewFromEnv() (*App, error) {
	// config.Load() validates every HTTP/DB/CORS/rate-limit setting up
	// front and fails fast with a clear error if something is missing or
	// out of range - a bad deploy should refuse to start, not start and
	// panic (or silently misbehave) on the first real request.
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	logger := newLogger(cfg.App.Environment, cfg.Log.Level)

	// database.Open (unlike a bare sql.Open) sets ConnMaxLifetime and
	// ConnMaxIdleTime. Without those, connections that MySQL has silently
	// dropped after its wait_timeout (or a stateful LB's idle timeout)
	// stay in the pool looking healthy until a request tries to use them,
	// producing intermittent "invalid connection" / "server has gone
	// away" errors in production. It also pings once at startup so a bad
	// DB_HOST/DB_USER/DB_PASSWORD fails immediately instead of on the
	// first real request.
	db, err := database.Open(context.Background(), cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	sqlDB := db.DB

	secret := []byte(env("JWT_SECRET", ""))
	if len(secret) < 32 {
		db.Close()
		return nil, fmt.Errorf("JWT_SECRET must be set to at least 32 random bytes (see .env.example)")
	}
	accessTTL, err := time.ParseDuration(env("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(env("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("JWT_REFRESH_TTL: %w", err)
	}
	maxFailedLogins, err := strconv.Atoi(env("AUTH_MAX_FAILED_REQUESTS", "10"))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("AUTH_MAX_FAILED_REQUESTS: %w", err)
	}

	// --- payroll calculator (public) ---
	rulesRepository := rr.Repository{DB: sqlDB}
	payeService := ps.Service{Rules: rulesRepository}
	minDate, _ := time.Parse("2006-01-02", "2022-01-01")
	payeHandler := ph.Handler{Service: payeService, MaxCustom: 20, MinDate: minDate}
	rulesListHandler := rh.Handler{Repo: rulesRepository}

	// --- customer auth + saved calculations (JWT-protected) ---
	companySvc := companyService.Service{Repo: companyRepo.Repository{DB: sqlDB}}
	companyH := companyHandler.Handler{Service: companySvc}
	departmentSvc := departmentService.Service{Repo: departmentRepo.Repository{DB: sqlDB}, CompanyRepo: companyRepo.Repository{DB: sqlDB}}
	departmentH := departmentHandler.Handler{Service: departmentSvc}
	employeeSvc := employeeService.Service{Repo: employeeRepo.Repository{DB: sqlDB}, CompanyRepo: companyRepo.Repository{DB: sqlDB}}
	employeeH := employeeHandler.Handler{Service: employeeSvc}
	payrollRunSvc := payrollRunService.Service{Repo: payrollRunRepo.Repository{DB: sqlDB}, CompanyRepo: companyRepo.Repository{DB: sqlDB}, Calculator: payeService}
	payrollRunH := payrollRunHandler.Handler{Service: payrollRunSvc}

	authSvc := authService.Service{Repo: authRepo.Repository{DB: sqlDB}, Secret: secret, AccessTTL: accessTTL, RefreshTTL: refreshTTL, MaxFailedLogins: maxFailedLogins}
	authH := authHandler.Handler{Service: authSvc}
	savedH := savedHandler.Handler{Repo: savedRepo.Repository{DB: sqlDB}}

	// --- admin auth + rule management (JWT-protected, permission-gated) ---
	adminSvc := adminService.Service{Repo: adminRepo.Repository{DB: sqlDB}, Secret: secret, AccessTTL: accessTTL, RefreshTTL: refreshTTL, MaxFailedLogins: maxFailedLogins}
	auditWriter := audit.Writer{DB: sqlDB}
	adminH := adminHandler.Handler{Service: adminSvc, Audit: auditWriter}
	rulesPersistRepo := rulesRepo.Repository{DB: sqlDB}
	rulesH := rulesHandler.Handler{Repo: rulesPersistRepo, Audit: auditWriter}
	workflowH := rulesHandler.WorkflowHandler{Repo: rulesPersistRepo, Workflow: rulesWorkflow.Workflow{DB: sqlDB}, Audit: auditWriter}

	// --- health (used by load balancers/orchestrators/uptime checks,
	// not by end users) ---
	healthH := health.Handler{DB: db, Version: cfg.App.Version}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthH.Liveness)
	mux.HandleFunc("GET /api/v1/ready", healthH.Readiness)
	mux.HandleFunc("POST /api/v1/calculator/paye", payeHandler.Calculate)
	mux.HandleFunc("GET /api/v1/payroll/rules", rulesListHandler.Applicable)

	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)
	mux.Handle("POST /api/v1/auth/change-password", middleware.RequireAuth(secret, http.HandlerFunc(authH.ChangePassword)))
	mux.Handle("POST /api/v1/calculations", middleware.RequireAuth(secret, http.HandlerFunc(savedH.Save)))
	mux.Handle("GET /api/v1/calculations", middleware.RequireAuth(secret, http.HandlerFunc(savedH.List)))
	mux.Handle("PATCH /api/v1/calculations/{id}", middleware.RequireAuth(secret, http.HandlerFunc(savedH.Rename)))
	mux.Handle("DELETE /api/v1/calculations/{id}", middleware.RequireAuth(secret, http.HandlerFunc(savedH.Delete)))

	// --- multi-company payroll foundation ---
	mux.Handle("POST /api/v1/companies", middleware.RequireAuth(secret, http.HandlerFunc(companyH.CreateCompany)))
	mux.Handle("GET /api/v1/companies", middleware.RequireAuth(secret, http.HandlerFunc(companyH.ListCompanies)))
	mux.Handle("GET /api/v1/companies/{company_id}", middleware.RequireAuth(secret, http.HandlerFunc(companyH.GetCompany)))
	mux.Handle("PATCH /api/v1/companies/{company_id}", middleware.RequireAuth(secret, http.HandlerFunc(companyH.UpdateCompany)))
	mux.Handle("GET /api/v1/companies/{company_id}/roles", middleware.RequireAuth(secret, http.HandlerFunc(companyH.ListRoles)))
	mux.Handle("POST /api/v1/companies/{company_id}/roles", middleware.RequireAuth(secret, http.HandlerFunc(companyH.CreateRole)))
	mux.Handle("GET /api/v1/companies/{company_id}/members", middleware.RequireAuth(secret, http.HandlerFunc(companyH.ListMembers)))
	mux.Handle("POST /api/v1/companies/{company_id}/members", middleware.RequireAuth(secret, http.HandlerFunc(companyH.AddMember)))
	mux.Handle("PATCH /api/v1/companies/{company_id}/members/{member_id}/role", middleware.RequireAuth(secret, http.HandlerFunc(companyH.UpdateMemberRole)))
	mux.Handle("GET /api/v1/companies/{company_id}/departments", middleware.RequireAuth(secret, http.HandlerFunc(departmentH.List)))
	mux.Handle("POST /api/v1/companies/{company_id}/departments", middleware.RequireAuth(secret, http.HandlerFunc(departmentH.Create)))
	mux.Handle("PATCH /api/v1/companies/{company_id}/departments/{department_id}", middleware.RequireAuth(secret, http.HandlerFunc(departmentH.Update)))
	mux.Handle("POST /api/v1/companies/{company_id}/employees", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.Create)))
	mux.Handle("GET /api/v1/companies/{company_id}/employees", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.List)))
	mux.Handle("GET /api/v1/companies/{company_id}/employees/{employee_id}", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.Get)))
	mux.Handle("PATCH /api/v1/companies/{company_id}/employees/{employee_id}", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.Update)))
	mux.Handle("GET /api/v1/companies/{company_id}/employees/{employee_id}/salary-history", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.SalaryHistory)))
	mux.Handle("POST /api/v1/companies/{company_id}/employees/{employee_id}/salary-history", middleware.RequireAuth(secret, http.HandlerFunc(employeeH.AddSalary)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Create)))
	mux.Handle("GET /api/v1/companies/{company_id}/payroll-runs", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.List)))
	mux.Handle("GET /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Get)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/calculate", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Calculate)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/employees/{employee_run_id}/adjustments", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.AddAdjustment)))
	mux.Handle("PATCH /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/employees/{employee_run_id}/adjustments/{adjustment_id}", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.UpdateAdjustment)))
	mux.Handle("DELETE /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/employees/{employee_run_id}/adjustments/{adjustment_id}", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.DeleteAdjustment)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/review", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Review)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/approve", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Approve)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/finalize", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Finalize)))
	mux.Handle("POST /api/v1/companies/{company_id}/payroll-runs/{payroll_run_id}/lock", middleware.RequireAuth(secret, http.HandlerFunc(payrollRunH.Lock)))

	mux.HandleFunc("POST /api/v1/admin/auth/login", adminH.Login)
	mux.HandleFunc("POST /api/v1/admin/auth/refresh", adminH.Refresh)
	mux.Handle("POST /api/v1/admin/auth/change-password", middleware.RequireAdmin(secret, http.HandlerFunc(adminH.ChangePassword)))
	mux.Handle("GET /api/v1/admin/live-rules", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.read", http.HandlerFunc(rulesListHandler.AdminDetail))))
	mux.Handle("GET /api/v1/admin/rule-sets", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.read", http.HandlerFunc(rulesH.List))))
	mux.Handle("POST /api/v1/admin/rule-sets", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.write", http.HandlerFunc(rulesH.Create))))
	mux.Handle("POST /api/v1/admin/rule-sets/validate", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.write", http.HandlerFunc(rulesH.Validate))))
	mux.Handle("POST /api/v1/admin/rule-sets/preview", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.read", http.HandlerFunc(rulesH.Preview))))
	mux.Handle("POST /api/v1/admin/rule-sets/{id}/submit-review", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.write", http.HandlerFunc(workflowH.SubmitReview))))
	mux.Handle("POST /api/v1/admin/rule-sets/{id}/approve", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.review", http.HandlerFunc(workflowH.Approve))))
	mux.Handle("POST /api/v1/admin/rule-sets/{id}/reject", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.review", http.HandlerFunc(workflowH.Reject))))
	mux.Handle("POST /api/v1/admin/rule-sets/{id}/publish", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.publish", http.HandlerFunc(workflowH.Publish))))
	mux.Handle("POST /api/v1/admin/rule-sets/{id}/archive", middleware.RequireAdmin(secret, middleware.RequirePermission("rules.write", http.HandlerFunc(workflowH.Archive))))

	mux.Handle("GET /api/v1/admin/users", middleware.RequireAdmin(secret, middleware.RequirePermission("admin.users", http.HandlerFunc(adminH.ListUsers))))
	mux.Handle("POST /api/v1/admin/users", middleware.RequireAdmin(secret, middleware.RequirePermission("admin.users", http.HandlerFunc(adminH.CreateUser))))
	mux.Handle("PATCH /api/v1/admin/users/{id}/status", middleware.RequireAdmin(secret, middleware.RequirePermission("admin.users", http.HandlerFunc(adminH.SetUserStatus))))

	mux.Handle("GET /api/v1/admin/audit-logs", middleware.RequireAdmin(secret, middleware.RequirePermission("audit.read", http.HandlerFunc(adminH.ListAuditLogs))))

	// CORS_ALLOWED_ORIGINS falls back to the admin frontend's Vite dev
	// origin so local dev keeps working with zero config; set it
	// explicitly in production (see .env.production.example).
	corsCfg := cfg.CORS
	if len(corsCfg.AllowedOrigins) == 0 {
		corsCfg.AllowedOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	// Order matters. Applied outer-to-inner below (each line wraps the
	// previous), which means execution order is the reverse of the list:
	// RequestID runs first (so every later stage, including Recovery, can
	// log/report it) -> Recovery wraps EVERYTHING downstream so a panic
	// anywhere in security headers, CORS, rate limiting, body limits, or
	// any handler is caught and turned into a clean 500 instead of
	// crashing the process or hanging the connection -> SecurityHeaders
	// -> CORS -> RateLimit -> MaxBodyBytes -> the route mux.
	var h http.Handler = mux
	h = middleware.MaxBodyBytes(cfg.HTTP.MaxRequestBodyBytes)(h)
	h = middleware.RateLimit(cfg.RateLimit, cfg.HTTP.TrustProxyHeaders)(h)
	h = middleware.CORS(corsCfg)(h)
	h = middleware.SecurityHeaders(h)
	h = middleware.Recovery(logger)(h)
	h = middleware.RequestID(h)

	server := &http.Server{
		Addr:              cfg.HTTP.Host + ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:           h,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	return &App{Server: server, DB: db, Log: logger, shutdownTimeout: cfg.HTTP.ShutdownTimeout}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.Log.Info("budget254-paye-api listening", "addr", a.Server.Addr)
	ch := make(chan error, 1)
	go func() { ch <- a.Server.ListenAndServe() }()
	select {
	case <-ctx.Done():
		a.Log.Info("shutting down")
		c, cancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
		defer cancel()
		err := a.Server.Shutdown(c)
		if closeErr := a.DB.Close(); closeErr != nil {
			a.Log.Error("closing database", "error", closeErr)
		}
		return err
	case e := <-ch:
		if e != nil && e != http.ErrServerClosed {
			a.Log.Error("server stopped unexpectedly", "error", e)
		}
		return e
	}
}

func newLogger(environment, level string) *slog.Logger {
	lvl := slog.LevelInfo
	switch strings.ToUpper(level) {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	}
	// JSON everywhere except local development, where a human is reading
	// the terminal in real time.
	if environment == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
