package rules

import (
	"github.com/alumasinde/budget254-paye-api/internal/rules/handler"
	"net/http"
)

func RegisterRoutes(m *http.ServeMux, h *handler.Handler) {
	m.HandleFunc("GET /api/v1/rules", h.Applicable)
}
