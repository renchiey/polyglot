package handlers

import (
	"net/http"

	"github.com/renchieyang/polyglot/server/internal/dict"
	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/lexaudit"
)

// SegmentHandler tokenizes Chinese text so the client can map a tapped
// character to its compound word (Assisted Noticing) and render per-word
// pinyin over a whole passage.
type SegmentHandler struct {
	Registry *lexaudit.Registry
	Dict     *dict.Dictionary
}

type segmentRequest struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

type token struct {
	Text   string `json:"text"`
	Pinyin string `json:"pinyin,omitempty"`
	CJK    bool   `json:"cjk"`
}

type segmentResponse struct {
	Tokens []token `json:"tokens"`
}

// Segment splits text into ordered tokens. CJK tokens found in the dictionary
// carry their pinyin; punctuation and spaces come back with cjk=false so the
// client can reconstruct the original text.
func (h *SegmentHandler) Segment(w http.ResponseWriter, r *http.Request) {
	var req segmentRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Text == "" {
		httputil.Error(w, http.StatusBadRequest, "text is required")
		return
	}

	if !h.Registry.Supported(req.Language) {
		httputil.Error(w, http.StatusBadRequest, "unsupported language")
		return
	}
	auditor, err := h.Registry.Get(req.Language)
	if err != nil {
		httputil.Error(w, http.StatusServiceUnavailable, "lexicon unavailable")
		return
	}

	raw := auditor.Segment(req.Text)
	tokens := make([]token, 0, len(raw))
	for _, t := range raw {
		tok := token{Text: t, CJK: lexaudit.IsCJK(t)}
		if tok.CJK && h.Dict != nil {
			if e, ok := h.Dict.Lookup(t); ok {
				tok.Pinyin = e.Pinyin
			}
		}
		tokens = append(tokens, tok)
	}
	httputil.JSON(w, http.StatusOK, segmentResponse{Tokens: tokens})
}
