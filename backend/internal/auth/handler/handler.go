package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/alumasinde/budget254-paye-api/internal/auth/service"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/response"
)

type Handler struct{ Service service.Service }

type registerRequest struct{ Email, Password, FirstName, LastName string }
type loginRequest struct{ Email, Password string }
type refreshRequest struct{ RefreshToken string `json:"refresh_token"` }
type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func decode(w http.ResponseWriter, r *http.Request, v any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var q registerRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	u, t, err := h.Service.Register(r.Context(), q.Email, q.Password, q.FirstName, q.LastName, r.UserAgent())
	if err != nil {
		response.Fail(w, 422, "REGISTRATION_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 201, map[string]any{"user": map[string]string{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName}, "tokens": t})
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
	response.JSON(w, 200, map[string]any{"user": map[string]string{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName}, "tokens": t})
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
	response.JSON(w, 200, map[string]any{"user": map[string]string{"id": u.PublicID, "email": u.Email, "first_name": u.FirstName, "last_name": u.LastName}, "tokens": t})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var q changePasswordRequest
	if err := decode(w, r, &q); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}
	if err := h.Service.ChangePassword(r.Context(), middleware.UserID(r.Context()), q.OldPassword, q.NewPassword); err != nil {
		response.Fail(w, 422, "CHANGE_PASSWORD_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	w.WriteHeader(204)
}
