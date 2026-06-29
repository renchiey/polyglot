package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/renchieyang/polyglot/server/internal/auth"
	"github.com/renchieyang/polyglot/server/internal/db/gen"
	"github.com/renchieyang/polyglot/server/internal/httputil"
)

// AuthHandler holds dependencies for auth-related endpoints.
type AuthHandler struct {
	Queries   *gen.Queries
	JWTSecret string
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string   `json:"token"`
	User  userView `json:"user"`
}

type userView struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

// Register creates a new user and returns a signed JWT.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httputil.Decode(r, &in); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Email == "" || len(in.Password) < 8 {
		httputil.Error(w, http.StatusBadRequest, "email required and password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), gen.CreateUserParams{
		Email:        in.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		if isUniqueViolation(err) {
			httputil.Error(w, http.StatusConflict, "email already registered")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "could not create user")
		return
	}

	h.respondWithToken(w, http.StatusCreated, user.ID, user.Email, user.CreatedAt)
}

// Login verifies credentials and returns a signed JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httputil.Decode(r, &in); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))

	user, err := h.Queries.GetUserByEmail(r.Context(), in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "login failed")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	h.respondWithToken(w, http.StatusOK, user.ID, user.Email, user.CreatedAt)
}

// Me returns the authenticated user's profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	idStr, ok := auth.UserID(r.Context())
	if !ok {
		httputil.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid token subject")
		return
	}

	user, err := h.Queries.GetUserByID(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusNotFound, "user not found")
		return
	}

	httputil.JSON(w, http.StatusOK, userView{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

func (h *AuthHandler) respondWithToken(w http.ResponseWriter, status int, id uuid.UUID, email string, createdAt time.Time) {
	token, err := auth.GenerateToken(h.JWTSecret, id.String())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	httputil.JSON(w, status, authResponse{
		Token: token,
		User:  userView{ID: id, Email: email, CreatedAt: createdAt},
	})
}

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "SQLSTATE 23505")
}
