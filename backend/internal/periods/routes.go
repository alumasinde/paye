package periods

import (
	"github.com/alumasinde/budget254-paye-api/internal/periods/handler"
	"net/http"
)

func RegisterRoutes(m *http.ServeMux, h handler.Handler) {
	m.HandleFunc("GET /api/v1/periods/resolve", h.Resolve)
}
