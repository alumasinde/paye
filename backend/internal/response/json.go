package response

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func JSON(w http.ResponseWriter, status int, p any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Never let a browser (or any intermediate cache) reuse a stale API
	// response - important for GET endpoints like /admin/rule-sets and
	// /admin/audit-logs, and critical for error bodies: without this, a
	// browser can keep serving an old cached error (e.g. a plain-text
	// body from before a bug fix) instead of re-checking the server.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(p)
}
func Error(w http.ResponseWriter, status int, code, msg, id string) {
	JSON(w, status, ErrorBody{code, msg, id})
}
