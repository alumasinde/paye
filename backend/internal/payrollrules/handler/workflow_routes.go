package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/alumasinde/budget254-paye-api/internal/admin/audit"
	"github.com/alumasinde/budget254-paye-api/internal/middleware"
	repo "github.com/alumasinde/budget254-paye-api/internal/payrollrules/repository"
	workflow "github.com/alumasinde/budget254-paye-api/internal/payrollrules/service"
	"github.com/alumasinde/budget254-paye-api/internal/response"
)

type WorkflowHandler struct {
	Repo     repo.Repository
	Workflow workflow.Workflow
	Audit    audit.Writer
}

type reviewRequest struct{ Comment string }

func (h WorkflowHandler) resolve(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := h.Repo.IDByPublicID(r.Context(), r.PathValue("id"))
	if err != nil {
		response.Fail(w, 404, "RULE_SET_NOT_FOUND", "rule set not found", middleware.ID(r.Context()), nil)
		return 0, false
	}
	return id, true
}

func (h WorkflowHandler) SubmitReview(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolve(w, r)
	if !ok {
		return
	}
	if err := h.Workflow.SubmitReview(r.Context(), id, middleware.AdminDBID(r.Context())); err != nil {
		response.Fail(w, 422, "SUBMIT_REVIEW_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]string{"status": "submitted for review"})
}

func (h WorkflowHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolve(w, r)
	if !ok {
		return
	}
	var q reviewRequest
	_ = json.NewDecoder(r.Body).Decode(&q)
	if err := h.Workflow.Review(r.Context(), id, middleware.AdminDBID(r.Context()), true, q.Comment); err != nil {
		response.Fail(w, 422, "APPROVE_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]string{"status": "approved"})
}

func (h WorkflowHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolve(w, r)
	if !ok {
		return
	}
	var q reviewRequest
	_ = json.NewDecoder(r.Body).Decode(&q)
	if err := h.Workflow.Review(r.Context(), id, middleware.AdminDBID(r.Context()), false, q.Comment); err != nil {
		response.Fail(w, 422, "REJECT_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]string{"status": "rejected"})
}

func (h WorkflowHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolve(w, r)
	if !ok {
		return
	}
	adminID := middleware.AdminDBID(r.Context())
	if err := h.Workflow.Publish(r.Context(), id, adminID); err != nil {
		response.Fail(w, 422, "PUBLISH_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	publicID := r.PathValue("id")
	if auditErr := h.Audit.Write(r.Context(), &adminID, "rule_set.publish", "payroll_rule_set", publicID, middleware.ID(r.Context()), r.RemoteAddr, nil, map[string]string{"status": "PUBLISHED"}); auditErr != nil {
		log.Printf("audit write failed (rule_set.publish %s): %v", publicID, auditErr)
	}
	response.JSON(w, 200, map[string]string{"status": "published"})
}

func (h WorkflowHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id, ok := h.resolve(w, r)
	if !ok {
		return
	}
	if err := h.Workflow.Archive(r.Context(), id); err != nil {
		response.Fail(w, 422, "ARCHIVE_FAILED", err.Error(), middleware.ID(r.Context()), nil)
		return
	}
	response.JSON(w, 200, map[string]string{"status": "archived"})
}
