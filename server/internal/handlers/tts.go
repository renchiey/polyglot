package handlers

import (
	"net/http"

	"github.com/renchieyang/polyglot/server/internal/httputil"
	"github.com/renchieyang/polyglot/server/internal/tts"
)

// maxTTSChars bounds synthesis input so a request can't tie up the CLI.
const maxTTSChars = 400

// Speed is clamped to this range before being turned into a Piper length scale,
// so a client can't request an absurdly slow (long-running) or fast synthesis.
const (
	minTTSSpeed = 0.5
	maxTTSSpeed = 2.0
)

// TTSHandler synthesizes Mandarin speech via Piper. When Piper isn't configured
// it returns 503 and the client falls back to browser speech synthesis.
type TTSHandler struct {
	Synth *tts.Synthesizer
}

type ttsRequest struct {
	Text string `json:"text"`
	// Speed is a playback-rate multiplier (1.0 = the voice's natural pace).
	// Omitted/zero means natural pace.
	Speed float64 `json:"speed"`
}

// Speak returns WAV audio for the given text.
func (h *TTSHandler) Speak(w http.ResponseWriter, r *http.Request) {
	if h.Synth == nil || !h.Synth.Available() {
		httputil.Error(w, http.StatusServiceUnavailable, "tts unavailable")
		return
	}

	var req ttsRequest
	if err := httputil.Decode(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Text == "" {
		httputil.Error(w, http.StatusBadRequest, "text is required")
		return
	}
	if runes := []rune(req.Text); len(runes) > maxTTSChars {
		req.Text = string(runes[:maxTTSChars])
	}

	// Piper's length scale is the inverse of playback speed: faster speech means
	// shorter phonemes. Zero leaves the voice at its natural pace.
	var lengthScale float64
	if req.Speed > 0 {
		speed := req.Speed
		if speed < minTTSSpeed {
			speed = minTTSSpeed
		} else if speed > maxTTSSpeed {
			speed = maxTTSSpeed
		}
		lengthScale = 1.0 / speed
	}

	audio, err := h.Synth.Synth(r.Context(), req.Text, lengthScale)
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, "synthesis failed")
		return
	}

	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio)
}
