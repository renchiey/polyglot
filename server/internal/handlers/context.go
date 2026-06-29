package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/renchieyang/polyglot/server/internal/auth"
	"github.com/renchieyang/polyglot/server/internal/httputil"
)

// userUUID extracts and parses the authenticated user's ID from the request
// context. On failure it writes a 401 and returns ok=false, so callers can
// simply `return` when ok is false.
func userUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	idStr, ok := auth.UserID(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid token subject")
		return uuid.Nil, false
	}
	return id, true
}
