package handler

import (
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	"github.com/alumasinde/budget254-paye-api/internal/periods/service"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	"net/http"
)

type Handler struct{}

func (h Handler) Resolve(w http.ResponseWriter, r *http.Request) {
	p, e := service.Parse(r.URL.Query().Get("date"))
	if e != nil {
		response.Error(w, 400, "INVALID_DATE", "date must be YYYY-MM-DD", middleware.RequestIDFromContext(r.Context()))
		return
	}
	response.JSON(w, 200, p)
}
