package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alumasinde/budget254-paye-api/internal/company/model"
	repo "github.com/alumasinde/budget254-paye-api/internal/company/repository"
	"github.com/alumasinde/budget254-paye-api/internal/company/service"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct{ Service service.Service }

type companyRequest struct {
	LegalName        string `json:"legal_name"`
	TradingName      string `json:"trading_name"`
	KRAPIN           string `json:"kra_pin"`
	Email            string `json:"email"`
	Phone            string `json:"phone"`
	CountryCode      string `json:"country_code"`
	CurrencyCode     string `json:"currency_code"`
	PayrollFrequency string `json:"payroll_frequency"`
}

type roleRequest struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type memberRequest struct {
	Email  string `json:"email"`
	RoleID string `json:"role_id"`
}

type memberRoleRequest struct {
	RoleID string `json:"role_id"`
}

func (h Handler) CreateCompany(w http.ResponseWriter, r *http.Request) {
	var q companyRequest
	if err := decode(w, r, &q); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	c, err := h.Service.CreateCompany(r.Context(), middleware.UserID(r.Context()), toCreate(q))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, c)
}

func (h Handler) ListCompanies(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.ListCompanies(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"companies": items})
}

func (h Handler) GetCompany(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("company_id")
	c, err := h.Service.Company(r.Context(), id, middleware.UserID(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (h Handler) UpdateCompany(w http.ResponseWriter, r *http.Request) {
	var q companyRequest
	if err := decode(w, r, &q); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	c, err := h.Service.UpdateCompany(r.Context(), r.PathValue("company_id"), middleware.UserID(r.Context()), model.UpdateCompanyInput{
		LegalName: q.LegalName, TradingName: q.TradingName, Email: q.Email, Phone: q.Phone,
		CountryCode: q.CountryCode, CurrencyCode: q.CurrencyCode, PayrollFrequency: q.PayrollFrequency,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, c)
}

func (h Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.ListRoles(r.Context(), r.PathValue("company_id"), middleware.UserID(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"roles": items})
}

func (h Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var q roleRequest
	if err := decode(w, r, &q); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	role, err := h.Service.CreateRole(r.Context(), r.PathValue("company_id"), middleware.UserID(r.Context()), model.CreateRoleInput{
		Code: q.Code, Name: q.Name, Description: q.Description, Permissions: q.Permissions,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, role)
}

func (h Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.ListMembers(r.Context(), r.PathValue("company_id"), middleware.UserID(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"members": items})
}

func (h Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	var q memberRequest
	if err := decode(w, r, &q); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	member, err := h.Service.AddMember(r.Context(), r.PathValue("company_id"), middleware.UserID(r.Context()), q.Email, q.RoleID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusCreated, member)
}

func (h Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var q memberRoleRequest
	if err := decode(w, r, &q); err != nil {
		fail(w, r, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}
	member, err := h.Service.UpdateMemberRole(
		r.Context(),
		r.PathValue("company_id"),
		middleware.UserID(r.Context()),
		r.PathValue("member_id"),
		q.RoleID,
	)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response.JSON(w, http.StatusOK, member)
}

func decode(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func toCreate(q companyRequest) model.CreateCompanyInput {
	return model.CreateCompanyInput{
		LegalName: q.LegalName, TradingName: q.TradingName, KRAPIN: q.KRAPIN,
		Email: q.Email, Phone: q.Phone, CountryCode: q.CountryCode,
		CurrencyCode: q.CurrencyCode, PayrollFrequency: q.PayrollFrequency,
	}
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, repo.ErrForbidden):
		fail(w, r, http.StatusForbidden, "FORBIDDEN", "you do not have permission for this action")
	case errors.Is(err, repo.ErrNotFound):
		fail(w, r, http.StatusNotFound, "NOT_FOUND", "requested resource was not found")
	case errors.Is(err, repo.ErrConflict):
		fail(w, r, http.StatusConflict, "CONFLICT", "a record with those details already exists")
	default:
		message := strings.TrimSpace(err.Error())
		if message == "" {
			message = "request could not be completed"
		}
		fail(w, r, http.StatusUnprocessableEntity, "REQUEST_FAILED", message)
	}
}

func fail(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	response.Fail(w, status, code, message, middleware.ID(r.Context()), nil)
}
