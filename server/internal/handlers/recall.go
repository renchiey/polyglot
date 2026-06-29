package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/dict"
	"github.com/renchieyang/polyglot/server/internal/generate"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/progress"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

// clozeBlank replaces the target word in the generated sentence for the
// fill-in-the-blank (active recall) prompt.
const clozeBlank = "＿＿"

// RecallHandler implements generative testing (the PRD's context-lock-breaker):
// it generates a NEW sentence that uses a due word together with only the
// learner's known vocabulary, so the brain recalls the word rather than the
// geometry of one memorized sentence.
type RecallHandler struct {
	Pipeline *generate.Pipeline
	Registry *lexaudit.Registry
	Queries  *gen.Queries
	Dict     *dict.Dictionary
	Tracker  *progress.Tracker
}

type recallRequest struct {
	WordID uuid.UUID `json:"word_id"`
}

type recallResponse struct {
	Word     string `json:"word"`
	Pinyin   string `json:"pinyin,omitempty"`
	Sentence string `json:"sentence"`
	Cloze    string `json:"cloze"`
	Passed   bool   `json:"passed"`
	Rounds   int    `json:"rounds"`
}

// Recall generates a quiz sentence for a due word.
func (h *RecallHandler) Recall(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var req recallRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	word, err := h.Queries.GetWord(r.Context(), gen.GetWordParams{ID: req.WordID, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "word not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not load word")
		return
	}

	// Filler vocabulary tracks the learner's level (Vocabulary Elo); the target
	// word and known words are exempt from the audit regardless.
	target, err := h.Tracker.RecommendedLevel(r.Context(), uid, hskMaxLevel)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load progress")
		return
	}
	auditor, ok := resolveAuditor(w, h.Registry, "zh", target)
	if !ok {
		return
	}

	known, err := h.Queries.ListKnownTerms(r.Context(), gen.ListKnownTermsParams{
		UserID: uid,
		State:  srs.ReviewState,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vocabulary")
		return
	}

	result, err := h.Pipeline.Run(r.Context(), auditor, generate.GenRequest{
		Language:    "zh",
		TargetLevel: target,
		Kind:        "single short sentence",
		KnownWords:  known,
		MustInclude: []string{word.Term},
	})
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "generation failed")
		return
	}

	resp := recallResponse{
		Word:     word.Term,
		Sentence: result.Text,
		Cloze:    strings.ReplaceAll(result.Text, word.Term, clozeBlank),
		Passed:   result.Passed,
		Rounds:   result.Rounds,
	}
	if h.Dict != nil {
		if e, ok := h.Dict.Lookup(word.Term); ok {
			resp.Pinyin = e.Pinyin
		}
	}
	httputil.JSON(w, http.StatusOK, resp)
}
