package handlers

import (
	"net/http"
	"strings"

	"github.com/renchieyang/polyglot/server/internal/dict"
	"github.com/renchieyang/polyglot/server/internal/httputil"
)

// LookupHandler serves Mandarin word lookups (pinyin, definitions, and
// per-character breakdown) for the Assisted Noticing UI.
type LookupHandler struct {
	Dict *dict.Dictionary
}

type lookupResponse struct {
	Word        string       `json:"word"`
	Pinyin      string       `json:"pinyin"`
	Definitions []string     `json:"definitions"`
	Characters  []dict.Entry `json:"characters,omitempty"`
}

// mandarinCodes are the language codes that map to the Mandarin dictionary.
var mandarinCodes = map[string]bool{"": true, "zh": true, "cmn": true, "mandarin": true}

// Lookup handles GET /lookup?word=…&language=zh. language is optional and
// defaults to Mandarin, the only language currently supported.
func (h *LookupHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	if h.Dict == nil {
		httputil.Error(w, http.StatusServiceUnavailable, "dictionary unavailable")
		return
	}

	if lang := r.URL.Query().Get("language"); !mandarinCodes[lang] {
		httputil.Error(w, http.StatusBadRequest, "unsupported language")
		return
	}

	word := strings.TrimSpace(r.URL.Query().Get("word"))
	if word == "" {
		httputil.Error(w, http.StatusBadRequest, "word is required")
		return
	}

	entry, ok := h.Dict.Lookup(word)
	if !ok {
		httputil.Error(w, http.StatusNotFound, "word not found")
		return
	}

	httputil.JSON(w, http.StatusOK, lookupResponse{
		Word:        entry.Word,
		Pinyin:      entry.Pinyin,
		Definitions: entry.Definitions,
		Characters:  h.Dict.Breakdown(word),
	})
}
