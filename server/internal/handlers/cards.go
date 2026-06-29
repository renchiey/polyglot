package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/progress"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

const (
	defaultDueLimit = 20
	maxDueLimit     = 100
)

// CardsHandler serves the spaced-repetition study queue and applies reviews.
type CardsHandler struct {
	Queries   *gen.Queries
	Scheduler *srs.Scheduler
	Tracker   *progress.Tracker
}

// cardView is the schedule state the client needs to render and track a card.
type cardView struct {
	WordID        uuid.UUID `json:"word_id"`
	Term          string    `json:"term"`
	Translation   string    `json:"translation"`
	Due           time.Time `json:"due"`
	State         int16     `json:"state"`
	Reps          int32     `json:"reps"`
	Lapses        int32     `json:"lapses"`
	ScheduledDays int64     `json:"scheduled_days"`
}

// Due returns cards due now, soonest first. Optional ?limit= (default 20, max 100).
func (h *CardsHandler) Due(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"))
	rows, err := h.Queries.ListDueCardsWithWord(r.Context(), gen.ListDueCardsWithWordParams{
		UserID: uid,
		Due:    time.Now(),
		Limit:  limit,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load due cards")
		return
	}

	views := make([]cardView, 0, len(rows))
	for _, row := range rows {
		views = append(views, cardView{
			WordID:        row.WordID,
			Term:          row.Term,
			Translation:   row.Translation,
			Due:           row.Due,
			State:         row.State,
			Reps:          row.Reps,
			Lapses:        row.Lapses,
			ScheduledDays: row.ScheduledDays,
		})
	}
	httputil.JSON(w, http.StatusOK, views)
}

type reviewRequest struct {
	WordID uuid.UUID `json:"word_id"`
	Rating int       `json:"rating"`
}

// Review applies an FSRS grade (1=Again..4=Easy) to a word's card and returns
// the updated schedule. A word without a card yet starts from a fresh one.
func (h *CardsHandler) Review(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var req reviewRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	grade, ok := srs.ParseRating(req.Rating)
	if !ok {
		httputil.Error(w, http.StatusBadRequest, "rating must be 1 (again), 2 (hard), 3 (good) or 4 (easy)")
		return
	}

	// Verify the word belongs to the learner before scheduling against it.
	word, err := h.Queries.GetWord(r.Context(), gen.GetWordParams{ID: req.WordID, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "word not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not load word")
		return
	}

	saved, err := applyReview(r.Context(), h.Queries, h.Scheduler, uid, req.WordID, grade)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not save review")
		return
	}

	// Nudge Vocabulary Elo from this graded recall. Best effort: a tracking
	// failure must not fail the review the learner just completed.
	_, _ = h.Tracker.RecordVocabReview(r.Context(), uid, word.Term, req.Rating)

	httputil.JSON(w, http.StatusOK, cardView{
		WordID:        saved.WordID,
		Term:          word.Term,
		Translation:   word.Translation,
		Due:           saved.Due,
		State:         saved.State,
		Reps:          saved.Reps,
		Lapses:        saved.Lapses,
		ScheduledDays: saved.ScheduledDays,
	})
}

func parseLimit(raw string) int32 {
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultDueLimit
	}
	if n > maxDueLimit {
		return maxDueLimit
	}
	return int32(n)
}
