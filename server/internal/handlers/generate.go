package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/generate"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/progress"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

// GenerateHandler runs the Lexical Auditor pipeline: it produces
// comprehensible-input text for the authenticated learner, biased toward their
// saved vocabulary and audited to stay within the target HSK level.
type GenerateHandler struct {
	Pipeline *generate.Pipeline
	Registry *lexaudit.Registry
	Queries  *gen.Queries
	Tracker  *progress.Tracker
}

type generateRequest struct {
	Language string `json:"language"`
	// TargetLevel is optional; when omitted (<=0) it is derived from the
	// learner's Vocabulary Elo so content tracks their level automatically.
	TargetLevel int    `json:"target_level"`
	Topic       string `json:"topic"`
	Kind        string `json:"kind"`
}

// Generate produces a passage at the requested (or Elo-derived) level and
// returns it with its audit report and the number of pipeline rounds it took.
func (h *GenerateHandler) Generate(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var req generateRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	auditor, target, ok := h.resolveLevel(w, r, uid, req.Language, req.TargetLevel)
	if !ok {
		return
	}

	terms, err := h.Queries.ListKnownTerms(r.Context(), gen.ListKnownTermsParams{
		UserID: uid,
		State:  srs.ReviewState,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vocabulary")
		return
	}

	result, err := h.Pipeline.Run(r.Context(), auditor, generate.GenRequest{
		Language:    req.Language,
		TargetLevel: target,
		Topic:       req.Topic,
		Kind:        req.Kind,
		KnownWords:  terms,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "generation failed")
		return
	}

	httputil.JSON(w, http.StatusOK, result)
}

// resolveLevel validates the language and returns the auditor plus an effective
// target level: the requested one, or — when target <= 0 — one derived from the
// learner's Vocabulary Elo. Writes the error response and returns ok=false on
// failure (400 unsupported language / out-of-range, 503 lexicon unavailable).
func (h *GenerateHandler) resolveLevel(w http.ResponseWriter, r *http.Request, uid uuid.UUID, language string, target int) (*lexaudit.Auditor, int, bool) {
	if !h.Registry.Supported(language) {
		httputil.Error(w, http.StatusBadRequest, "unsupported language")
		return nil, 0, false
	}
	auditor, err := h.Registry.Get(language)
	if err != nil {
		httputil.Error(w, http.StatusServiceUnavailable, "lexicon unavailable")
		return nil, 0, false
	}

	if target <= 0 {
		target, err = h.Tracker.RecommendedLevel(r.Context(), uid, auditor.MaxLevel())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "could not load progress")
			return nil, 0, false
		}
	}
	if target < 1 || target > auditor.MaxLevel() {
		httputil.Error(w, http.StatusBadRequest, "target_level out of range")
		return nil, 0, false
	}
	return auditor, target, true
}
