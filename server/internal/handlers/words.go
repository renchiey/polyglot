package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/srs"
)

// WordsHandler manages the learner's vocabulary vault. Creating a word also
// seeds an FSRS card so it enters the study queue immediately.
type WordsHandler struct {
	Queries   *gen.Queries
	Scheduler *srs.Scheduler
}

type wordInput struct {
	Term        string `json:"term"`
	Translation string `json:"translation"`
	Definition  string `json:"definition"`
}

type wordView struct {
	ID          uuid.UUID  `json:"id"`
	Term        string     `json:"term"`
	Translation string     `json:"translation"`
	Definition  string     `json:"definition"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	NextReview  *time.Time `json:"next_review,omitempty"`
}

func toWordView(w gen.Word) wordView {
	return wordView{
		ID:          w.ID,
		Term:        w.Term,
		Translation: w.Translation,
		Definition:  w.Definition,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

// Create adds a word to the vault and seeds its spaced-repetition card.
func (h *WordsHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var in wordInput
	if err := httputil.Decode(r, &in); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.Term = strings.TrimSpace(in.Term)
	if in.Term == "" {
		httputil.Error(w, http.StatusBadRequest, "term is required")
		return
	}

	// Re-adding a word already in the vault means the learner forgot it: reset
	// that word's card to a fresh schedule rather than creating a duplicate.
	if existing, err := h.Queries.GetWordByTerm(r.Context(), gen.GetWordByTermParams{UserID: uid, Term: in.Term}); err == nil {
		if _, err := h.Queries.UpsertCard(r.Context(), srs.ToUpsertParams(uid, existing.ID, srs.NewCard(time.Now()))); err != nil {
			httputil.Error(w, http.StatusInternalServerError, "could not reset word")
			return
		}
		httputil.JSON(w, http.StatusOK, toWordView(existing))
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		httputil.Error(w, http.StatusInternalServerError, "could not check vault")
		return
	}

	word, err := h.Queries.CreateWord(r.Context(), gen.CreateWordParams{
		UserID:      uid,
		Term:        in.Term,
		Translation: in.Translation,
		Definition:  in.Definition,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not create word")
		return
	}

	// Seed a fresh card so the word is immediately due for study.
	if _, err := h.Queries.UpsertCard(r.Context(), srs.ToUpsertParams(uid, word.ID, srs.NewCard(time.Now()))); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not schedule word")
		return
	}

	httputil.JSON(w, http.StatusCreated, toWordView(word))
}

// List returns the learner's vault, newest first.
func (h *WordsHandler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	words, err := h.Queries.ListWordsWithCards(r.Context(), uid)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load words")
		return
	}

	views := make([]wordView, 0, len(words))
	for _, word := range words {
		views = append(views, wordView{
			ID:          word.ID,
			Term:        word.Term,
			Translation: word.Translation,
			Definition:  word.Definition,
			CreatedAt:   word.CreatedAt,
			UpdatedAt:   word.UpdatedAt,
			NextReview:  word.NextReview,
		})
	}
	httputil.JSON(w, http.StatusOK, views)
}

// Get returns a single word the learner owns.
func (h *WordsHandler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}

	word, err := h.Queries.GetWord(r.Context(), gen.GetWordParams{ID: id, UserID: uid})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "word not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not load word")
		return
	}
	httputil.JSON(w, http.StatusOK, toWordView(word))
}

// Update edits a word the learner owns.
func (h *WordsHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}

	var in wordInput
	if err := httputil.Decode(r, &in); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.Term = strings.TrimSpace(in.Term)
	if in.Term == "" {
		httputil.Error(w, http.StatusBadRequest, "term is required")
		return
	}

	word, err := h.Queries.UpdateWord(r.Context(), gen.UpdateWordParams{
		ID:          id,
		UserID:      uid,
		Term:        in.Term,
		Translation: in.Translation,
		Definition:  in.Definition,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "word not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not update word")
		return
	}
	httputil.JSON(w, http.StatusOK, toWordView(word))
}

// Delete removes a word and (via ON DELETE CASCADE) its card.
func (h *WordsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}

	if err := h.Queries.DeleteWord(r.Context(), gen.DeleteWordParams{ID: id, UserID: uid}); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not delete word")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// pathUUID parses a UUID URL parameter, writing a 400 on failure.
func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid "+name)
		return uuid.Nil, false
	}
	return id, true
}
