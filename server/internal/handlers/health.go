package handlers

import (
	"net/http"

	"github.com/renchieyang/polyglot/server/internal/httputil"
)

// Health is a simple liveness/readiness probe.
func Health(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
