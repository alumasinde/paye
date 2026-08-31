package handler

import (
	"net/http"
	"time"

	"github.com/alumasinde/budget254-paye-api/internal/payroll/dto"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	"github.com/alumasinde/budget254-paye-api/internal/rules/repository"
)

type Handler struct{ Repo repository.Repository }

func (h Handler) Applicable(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", d)
	if err != nil {
		response.JSON(w, 400, map[string]any{"code": "VALIDATION_ERROR", "message": "date must be YYYY-MM-DD"})
		return
	}

	rules, err := h.Repo.Applicable(r.Context(), date)
	if err != nil {
		response.JSON(w, 500, map[string]any{"code": "RULES_UNAVAILABLE", "message": "rules could not be resolved"})
		return
	}

	out := dto.RulesResponse{CalculationDate: d}
	for _, x := range rules {
		var effectiveTo *string
		if x.EffectiveTo != nil {
			value := x.EffectiveTo.Format("2006-01-02")
			effectiveTo = &value
		}

		out.Rules = append(out.Rules, dto.Rule{
			Code:          x.Code,
			Name:          x.Name,
			Version:       x.Version,
			Method:        x.Method,
			EffectiveFrom: x.EffectiveFrom.Format("2006-01-02"),
			EffectiveTo:   effectiveTo,
		})
	}

	response.JSON(w, 200, out)
}

// AdminDetail exposes fully-resolved rule data, including parameters and bands,
// to the admin panel for editing existing rules.
func (h Handler) AdminDetail(w http.ResponseWriter, r *http.Request) {
	d := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", d)
	if err != nil {
		response.JSON(w, 400, map[string]any{"code": "VALIDATION_ERROR", "message": "date must be YYYY-MM-DD"})
		return
	}

	rules, err := h.Repo.ResolvedApplicable(r.Context(), date)
	if err != nil {
		response.JSON(w, 500, map[string]any{"code": "RULES_UNAVAILABLE", "message": "rules could not be resolved"})
		return
	}

	response.JSON(w, 200, map[string]any{"calculation_date": d, "rules": rules})
}
