package handlers

import (
	"math"
	"net/http"

	"github.com/renchieyang/polyglot/server/internal/elo"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/progress"
)

// hskMaxLevel is the fallback ceiling for the recommended level when the
// Mandarin lexicon isn't loaded yet.
const hskMaxLevel = 7

// ProgressHandler exposes the learner's Linguistic Elo across the four PRD
// vectors plus a recommended HSK level derived from Vocabulary Elo.
type ProgressHandler struct {
	Tracker  *progress.Tracker
	Registry *lexaudit.Registry
}

type progressResponse struct {
	Vocabulary       int `json:"vocabulary"`
	Syntax           int `json:"syntax"`
	Listening        int `json:"listening"`
	Speaking         int `json:"speaking"`
	RecommendedLevel int `json:"recommended_level"`
}

// Progress returns the authenticated learner's ratings (rounded for display).
func (h *ProgressHandler) Progress(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	row, err := h.Tracker.Ratings(r.Context(), uid)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not load progress")
		return
	}

	maxLevel := hskMaxLevel
	if aud, err := h.Registry.Get("zh"); err == nil {
		maxLevel = aud.MaxLevel()
	}

	httputil.JSON(w, http.StatusOK, progressResponse{
		Vocabulary:       int(math.Round(row.Vocabulary)),
		Syntax:           int(math.Round(row.Syntax)),
		Listening:        int(math.Round(row.Listening)),
		Speaking:         int(math.Round(row.Speaking)),
		RecommendedLevel: elo.LevelForRating(row.Vocabulary, maxLevel),
	})
}
