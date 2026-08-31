package payroll

import (
	"github.com/alumasinde/budget254-paye-api/internal/payroll/handler"
	"net/http"
)

func Register(m *http.ServeMux, h handler.Handler) {
	m.HandleFunc("POST /api/v1/calculator/paye", h.Calculate)
}
