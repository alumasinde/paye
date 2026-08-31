package health

import (
	"context"
	"github.com/alumasinde/budget254-paye-api/internal/database"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	"net/http"
	"time"
)

type Handler struct {
	DB      *database.MySQL
	Version string
}

func (h Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, 200, map[string]any{"status": "ok", "service": "budget254-paye-api", "version": h.Version})
}
func (h Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	c, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if e := h.DB.Ping(c); e != nil {
		response.Error(w, 503, "DATABASE_UNAVAILABLE", "service dependency is unavailable", middleware.RequestIDFromContext(r.Context()))
		return
	}
	response.JSON(w, 200, map[string]any{"status": "ready", "database": "ok"})
}
