package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/alumasinde/budget254-paye-api/internal/admin/audit"
	"github.com/alumasinde/budget254-paye-api/internal/admin/service"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct {
	Service service.Service
	Audit   audit.Writer
}

type loginRequest struct{ Email, Password string }
type refreshRequest struct{ RefreshToken string `json:"refresh_token"` }
type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
type createAdminRequest struct {
	Email, Password, FirstName, LastName, Role string
}
type setStatusRequest struct{ Status string }

func decode(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var q loginRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	u, t, err := h.Service.Login(r.Context(), strings.TrimSpace(q.Email), q.Password, r.UserAgent())
	if err != nil {
		status, code := 401, "INVALID_CREDENTIALS"
		if errors.Is(err, service.ErrLocked) {
			status, code = 423, "ACCOUNT_LOCKED"
		}
		response.Fail(w, status, code, err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]any{
		"admin_user": map[string]any{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName, "roles": u.Roles, "permissions": u.Permissions},
		"tokens":     t,
	})
}

func (h Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var q refreshRequest
	if err := decode(w, r, &q); err != nil || strings.TrimSpace(q.RefreshToken) == "" {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	u, t, err := h.Service.Refresh(r.Context(), q.RefreshToken, r.UserAgent())
	if err != nil {
		response.Fail(w, 401, "INVALID_REFRESH_TOKEN", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]any{
		"admin_user": map[string]any{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName, "roles": u.Roles, "permissions": u.Permissions},
		"tokens":     t,
	})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var q changePasswordRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	if err := h.Service.ChangePassword(r.Context(), middleware.AdminID(r.Context()), q.OldPassword, q.NewPassword); err != nil {
		response.Fail(w, 422, "CHANGE_PASSWORD_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	w.WriteHeader(204)
}

func (h Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Service.List(r.Context())
	if err != nil {
		response.Fail(w, 500, "ADMIN_LIST_FAILED", "could not load admin users", middleware.ID(r.Context()), nil)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, u := range items {
		out = append(out, map[string]any{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName, "status": u.Status, "roles": u.Roles, "created_at": u.CreatedAt})
	}
	response.JSON(w, 200, map[string]any{"items": out})
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var q createAdminRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	u, err := h.Service.CreateAdmin(r.Context(), q.Email, q.Password, q.FirstName, q.LastName, q.Role)
	if err != nil {
		response.Fail(w, 422, "ADMIN_CREATE_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 201, map[string]any{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName, "status": u.Status, "roles": u.Roles})
}

func (h Handler) SetUserStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var q setStatusRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	if err := h.Service.SetStatus(r.Context(), id, strings.ToUpper(q.Status)); err != nil {
		response.Fail(w, 422, "STATUS_UPDATE_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	w.WriteHeader(204)
}

func (h Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 200 {
		limit = 50
	}
	items, err := h.Audit.List(r.Context(), limit)
	if err != nil {
		response.Fail(w, 500, "AUDIT_LIST_FAILED", "could not load audit log", middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]any{"items": items})
}
