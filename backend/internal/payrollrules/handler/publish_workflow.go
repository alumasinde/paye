package handler

import (
	"encoding/json"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	wf "github.com/alumasinde/budget254-paye-api/internal/payrollrules/service"
	"github.com/alumasinde/budget254-paye-api/internal/response"
	"net/http"
	"strings"
)

func workflowID(r *http.Request) string { return r.PathValue("id") }
func (h Handler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	id, err := h.Repo.InternalID(r.Context(), workflowID(r))
	if err != nil {
		response.JSON(w, 404, map[string]string{"message": "rule set not found"})
		return
	}
	if err := (wf.Workflow{DB: h.Repo.DB}).SubmitReview(r.Context(), id, middleware.AdminDBID(r.Context())); err != nil {
		response.JSON(w, 422, map[string]string{"message": err.Error()})
		return
	}
	response.JSON(w, 200, map[string]string{"message": "submitted for review"})
}
func (h Handler) Approve(w http.ResponseWriter, r *http.Request) { h.review(w, r, true) }
func (h Handler) Reject(w http.ResponseWriter, r *http.Request)  { h.review(w, r, false) }
func (h Handler) review(w http.ResponseWriter, r *http.Request, approve bool) {
	var body struct {
		Comment string `json:"comment"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body)
	id, err := h.Repo.InternalID(r.Context(), workflowID(r))
	if err != nil {
		response.JSON(w, 404, map[string]string{"message": "rule set not found"})
		return
	}
	if err := (wf.Workflow{DB: h.Repo.DB}).Review(r.Context(), id, middleware.AdminDBID(r.Context()), approve, strings.TrimSpace(body.Comment)); err != nil {
		response.JSON(w, 422, map[string]string{"message": err.Error()})
		return
	}
	response.JSON(w, 200, map[string]string{"message": map[bool]string{true: "approved", false: "rejected"}[approve]})
}
func (h Handler) Publish(w http.ResponseWriter, r *http.Request) {
	if err := h.Repo.PublishToLive(r.Context(), workflowID(r), middleware.AdminDBID(r.Context())); err != nil {
		response.JSON(w, 422, map[string]string{"message": err.Error()})
		return
	}
	response.JSON(w, 200, map[string]string{"message": "published to live calculator rules"})
}
func (h Handler) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := h.Repo.InternalID(r.Context(), workflowID(r))
	if err != nil {
		response.JSON(w, 404, map[string]string{"message": "rule set not found"})
		return
	}
	if err := (wf.Workflow{DB: h.Repo.DB}).Archive(r.Context(), id); err != nil {
		response.JSON(w, 422, map[string]string{"message": err.Error()})
		return
	}
	response.JSON(w, 200, map[string]string{"message": "archived"})
}
