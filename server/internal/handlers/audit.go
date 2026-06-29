package handlers

import (
	"net/http"

	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
)

// AuditHandler grades AI-generated sentences against a learner's level.
type AuditHandler struct {
	Registry *lexaudit.Registry
}

type auditRequest struct {
	Language    string `json:"language"`
	Sentence    string `json:"sentence"`
	TargetLevel int    `json:"target_level"`
}

type auditBatchRequest struct {
	Language    string   `json:"language"`
	Sentences   []string `json:"sentences"`
	TargetLevel int      `json:"target_level"`
}

// Audit grades a single sentence.
func (h *AuditHandler) Audit(w http.ResponseWriter, r *http.Request) {
	var req auditRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	auditor, ok := resolveAuditor(w, h.Registry, req.Language, req.TargetLevel)
	if !ok {
		return
	}
	httputil.JSON(w, http.StatusOK, auditor.Audit(req.Sentence, req.TargetLevel))
}

// AuditBatch grades several sentences at the same level in one call.
func (h *AuditHandler) AuditBatch(w http.ResponseWriter, r *http.Request) {
	var req auditBatchRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	auditor, ok := resolveAuditor(w, h.Registry, req.Language, req.TargetLevel)
	if !ok {
		return
	}

	reports := make([]lexaudit.Report, 0, len(req.Sentences))
	for _, s := range req.Sentences {
		reports = append(reports, auditor.Audit(s, req.TargetLevel))
	}
	httputil.JSON(w, http.StatusOK, reports)
}

// resolveAuditor validates the language and target level and returns the
// auditor for the requested language. It writes the appropriate error response
// and returns ok=false on failure: 400 for an unsupported language or
// out-of-range level, 503 when the lexicon is not yet loaded (e.g. first-run
// download still in progress or failed). Shared by the /audit and /generate
// handlers.
func resolveAuditor(w http.ResponseWriter, registry *lexaudit.Registry, language string, target int) (*lexaudit.Auditor, bool) {
	if !registry.Supported(language) {
		httputil.Error(w, http.StatusBadRequest, "unsupported language")
		return nil, false
	}

	auditor, err := registry.Get(language)
	if err != nil {
		httputil.Error(w, http.StatusServiceUnavailable, "lexicon unavailable")
		return nil, false
	}

	if target < 1 || target > auditor.MaxLevel() {
		httputil.Error(w, http.StatusBadRequest, "target_level out of range")
		return nil, false
	}
	return auditor, true
}
