package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
	"github.com/renchieyang/polyglot/server/internal/llm"
	"github.com/renchieyang/polyglot/server/internal/progress"
)

// AssessHandler grades a learner's English translation of a Chinese passage to
// gauge reading comprehension, and nudges Vocabulary Elo by the result.
type AssessHandler struct {
	LLM      llm.Client
	Registry *lexaudit.Registry
	Tracker  *progress.Tracker
}

type assessRequest struct {
	Text        string `json:"text"`
	Translation string `json:"translation"`
}

type assessResponse struct {
	Score    int    `json:"score"` // 0-100 comprehension
	Feedback string `json:"feedback"`
}

const assessSystem = `You grade a learner's English translation of a Chinese passage for COMPREHENSION (did they understand the meaning?), not style. Be encouraging but honest.

Output ONLY JSON: {"score": <0-100>, "feedback": "<one short sentence>"}.`

// Translation grades req.Translation against req.Text.
func (h *AssessHandler) Translation(w http.ResponseWriter, r *http.Request) {
	uid, ok := userUUID(w, r)
	if !ok {
		return
	}

	var req assessRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Text = strings.TrimSpace(req.Text)
	req.Translation = strings.TrimSpace(req.Translation)
	if req.Text == "" || req.Translation == "" {
		httputil.Error(w, http.StatusBadRequest, "text and translation are required")
		return
	}

	prompt := "Chinese passage:\n" + req.Text + "\n\nLearner's English translation:\n" + req.Translation
	resp, err := h.LLM.Complete(r.Context(), llm.Request{
		System:      assessSystem,
		Messages:    []llm.Message{{Role: llm.RoleUser, Content: prompt}},
		Temperature: 0.2,
		MaxTokens:   200,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "assessment failed")
		return
	}

	out := parseAssessment(resp.Text)

	// Nudge Vocabulary Elo: difficulty from the passage's highest HSK word.
	level := 1
	if aud, err := h.Registry.Get("zh"); err == nil {
		if rep := aud.Audit(req.Text, aud.MaxLevel()); rep.SentenceLevel > 0 {
			level = rep.SentenceLevel
		}
	}
	_, _ = h.Tracker.RecordReading(r.Context(), uid, level, float64(out.Score)/100.0)

	httputil.JSON(w, http.StatusOK, out)
}

// parseAssessment extracts the JSON object from the model's reply, tolerating
// surrounding prose. Falls back to a neutral result.
func parseAssessment(text string) assessResponse {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		var out assessResponse
		if err := json.Unmarshal([]byte(text[start:end+1]), &out); err == nil {
			if out.Score < 0 {
				out.Score = 0
			}
			if out.Score > 100 {
				out.Score = 100
			}
			if out.Feedback == "" {
				out.Feedback = "Assessment recorded."
			}
			return out
		}
	}
	return assessResponse{Score: 0, Feedback: "Couldn't grade that — try again."}
}
