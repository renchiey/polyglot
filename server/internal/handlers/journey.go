package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/journey"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/progress"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

// JourneyHandler generates and persists the Daily Journey story. Marking
// unknowns, learning them (add to vault) and recalling them happen client-side
// over /lookup, /words, /recall and /cards/review.
type JourneyHandler struct {
	Builder  *journey.Builder
	Registry *lexaudit.Registry
	Queries  *gen.Queries
	Tracker  *progress.Tracker
}

type journeyStartRequest struct {
	Topic       string `json:"topic"`
	TargetLevel int    `json:"target_level"`
}

type journeyView struct {
	ID          uuid.UUID            `json:"id"`
	Topic       string               `json:"topic"`
	Level       int32                `json:"level"`
	Passed      bool                 `json:"passed"`
	Story       []journey.Token      `json:"story"`
	Suggestions []journey.Suggestion `json:"suggestions"`
	Interaction journey.Interaction  `json:"interaction"`
}

func toJourneyView(row gen.Journey, c journey.Content) journeyView {
	return journeyView{
		ID:          row.ID,
		Topic:       row.Topic,
		Level:       row.Level,
		Passed:      c.Passed,
		Story:       c.Story,
		Suggestions: c.Suggestions,
		Interaction: c.Interaction,
	}
}

// Start generates an i+1 story at the learner's level and persists it.
func (h *JourneyHandler) Start(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var req journeyStartRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		topic = "everyday life"
	}

	auditor, err := h.Registry.Get("zh")
	if err != nil {
		httputil.Error(w, http.StatusServiceUnavailable, "lexicon unavailable")
		return
	}
	level := req.TargetLevel
	if level <= 0 {
		level, err = h.Tracker.RecommendedLevel(r.Context(), uid, auditor.MaxLevel())
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, "could not load progress")
			return
		}
	}
	if level < 1 || level > auditor.MaxLevel() {
		httputil.Error(w, http.StatusBadRequest, "target_level out of range")
		return
	}

	known, err := h.Queries.ListKnownTerms(r.Context(), gen.ListKnownTermsParams{UserID: uid, State: srs.ReviewState})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load vocabulary")
		return
	}

	content, err := h.Builder.Build(r.Context(), auditor, level, topic, known)
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "could not build journey")
		return
	}

	payload, err := json.Marshal(content)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not encode journey")
		return
	}
	row, err := h.Queries.CreateJourney(r.Context(), gen.CreateJourneyParams{
		UserID:  uid,
		Topic:   topic,
		Level:   int32(level),
		Content: payload,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not save journey")
		return
	}

	httputil.JSON(w, http.StatusCreated, toJourneyView(row, content))
}

// Get returns a persisted journey story.
func (h *JourneyHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}

	row, err := h.Queries.GetJourney(r.Context(), gen.GetJourneyParams{ID: id, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "journey not found")
		} else {
			httputil.Error(w, http.StatusInternalServerError, "could not load journey")
		}
		return
	}
	var content journey.Content
	if err := json.Unmarshal(row.Content, &content); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not decode journey")
		return
	}
	httputil.JSON(w, http.StatusOK, toJourneyView(row, content))
}
