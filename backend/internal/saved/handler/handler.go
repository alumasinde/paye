package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	repo "github.com/alumasinde/budget254-paye-api/internal/saved/repository"
)

type Handler struct{ Repo repo.Repository }

const maxLabelLength = 80

func (h Handler) Save(w http.ResponseWriter, r *http.Request) {
	var p map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&p); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid saved calculation", middleware.ID(r.Context()), nil)
		return
	}

	raw, _ := json.Marshal(p)
	s := func(k string) string { v, _ := p[k].(string); return v }
	stat, _ := json.Marshal(p["statutory_deductions"])
	custom, _ := json.Marshal(p["custom_deductions"])
	rules, _ := json.Marshal(p["rule_versions"])

	label, labelErr := normaliseLabel(p["label"])
	if labelErr != nil {
		response.Fail(w, 422, "INVALID_LABEL", labelErr.Error(), middleware.ID(r.Context()), nil)
		return
	}

	id, err := h.Repo.Save(r.Context(), middleware.UserID(r.Context()), label,
		s("calculation_date"), s("gross_salary"), s("taxable_income"), s("paye_before_relief"), s("relief"), s("paye"),
		s("total_deductions"), s("net_salary"), stat, custom, rules, raw)
	if err != nil {
		response.Fail(w, 500, "SAVE_FAILED", "could not save calculation", middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 201, map[string]string{"id": id})
}

func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	x, err := h.Repo.List(r.Context(), middleware.UserID(r.Context()), limit)
	if err != nil {
		response.Fail(w, 500, "HISTORY_FAILED", "could not load history", middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]any{"items": x})
}

func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.Repo.Delete(r.Context(), middleware.UserID(r.Context()), id); err != nil {
		response.Fail(w, 500, "DELETE_FAILED", "could not delete calculation", middleware.ID(r.Context()), nil)
		return
	}
	w.WriteHeader(204)
}

// Rename updates just the label on a saved calculation (e.g. renaming
// "Calculation for 2026-08-31" to "Job Offer A"). An empty/omitted label
// clears it back to no label, which the frontend falls back to showing
// the date for.
func (h Handler) Rename(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p struct {
		Label *string `json:"label"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&p); err != nil {
		response.Fail(w, 400, "INVALID_JSON", "invalid request", middleware.ID(r.Context()), nil)
		return
	}

	var labelValue *string
	if p.Label != nil {
		label, err := normaliseLabel(*p.Label)
		if err != nil {
			response.Fail(w, 422, "INVALID_LABEL", err.Error(), middleware.ID(r.Context()), nil)
			return
		}
		labelValue = label
	}

	if err := h.Repo.Rename(r.Context(), middleware.UserID(r.Context()), id, labelValue); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.Fail(w, 404, "CALCULATION_NOT_FOUND", "saved calculation not found", middleware.ID(r.Context()), nil)
			return
		}
		response.Fail(w, 500, "RENAME_FAILED", "could not rename calculation", middleware.ID(r.Context()), nil)
		return
	}
	w.WriteHeader(204)
}

// normaliseLabel trims whitespace, enforces the max length, and turns an
// empty string into nil (no label) rather than storing empty strings.
func normaliseLabel(raw any) (*string, error) {
	s, ok := raw.(string)
	if !ok {
		return nil, nil
	}
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, nil
	}
	if len(trimmed) > maxLabelLength {
		return nil, errors.New("label must be 80 characters or fewer")
	}
	return &trimmed, nil
}
