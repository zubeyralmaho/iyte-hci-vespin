package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
	tokens  *Tokens
}

func NewHandler(q *db.Queries, pool *pgxpool.Pool, tokens *Tokens) *Handler {
	return &Handler{queries: q, pool: pool, tokens: tokens}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/guest", h.createGuest)
	r.Post("/register", h.register)
	r.Post("/login", h.login)
}

func (h *Handler) createGuest(w http.ResponseWriter, r *http.Request) {
	user, err := h.queries.CreateGuestUser(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "create guest failed", "error", err)
		httpx.WriteError(w, fmt.Errorf("create guest: %w", err))
		return
	}

	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	token, err := h.tokens.Sign(userID, RoleGuest)
	if err != nil {
		slog.ErrorContext(r.Context(), "sign guest token failed", "error", err, "user_id", userID)
		httpx.WriteError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "guest created", "user_id", userID)
	httpx.WriteJSON(w, http.StatusCreated, AuthResponse{
		User:  ToUserResponse(user),
		Token: token,
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	email := normalizeEmail(req.Email)
	user, err := h.queries.GetUserByEmail(r.Context(), pgtype.Text{String: email, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		// The partial unique index on email guarantees this row, if found, has
		// role='registered'; no extra role check needed.
		httpx.WriteError(w, httpx.ErrInvalidCredentials)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "login lookup failed", "error", err)
		httpx.WriteError(w, fmt.Errorf("get user by email: %w", err))
		return
	}

	if err := VerifyPassword(user.PasswordHash.String, req.Password); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	token, err := h.tokens.Sign(userID, RoleRegistered)
	if err != nil {
		slog.ErrorContext(r.Context(), "sign login token failed", "error", err, "user_id", userID)
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, AuthResponse{
		User:  ToUserResponse(user),
		Token: token,
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	email := normalizeEmail(req.Email)
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		slog.ErrorContext(r.Context(), "hash password failed", "error", err)
		httpx.WriteError(w, err)
		return
	}

	guestID, mode, err := h.inspectAuthHeader(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	switch mode {
	case registerModeFresh:
		h.registerFresh(w, r, email, passwordHash, req.DisplayName)
	case registerModeConvert:
		h.registerConvert(w, r, guestID, email, passwordHash, req.DisplayName)
	case registerModeAlreadyRegistered:
		httpx.WriteError(w, httpx.ErrAlreadyRegistered)
	}
}

type registerMode int

const (
	registerModeFresh registerMode = iota
	registerModeConvert
	registerModeAlreadyRegistered
)

// inspectAuthHeader decides which branch /auth/register should take based on
// the Authorization header. Absent header → fresh register. Valid guest JWT
// → convert. Valid registered JWT → already-registered. Invalid token →
// ErrUnauthorized.
func (h *Handler) inspectAuthHeader(r *http.Request) (uuid.UUID, registerMode, error) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return uuid.Nil, registerModeFresh, nil
	}

	tokenValue, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(tokenValue) == "" {
		return uuid.Nil, 0, httpx.ErrUnauthorized
	}

	claims, err := h.tokens.Verify(strings.TrimSpace(tokenValue))
	if err != nil {
		return uuid.Nil, 0, err
	}

	switch claims.Role {
	case RoleGuest:
		return claims.UserID, registerModeConvert, nil
	case RoleRegistered:
		return uuid.Nil, registerModeAlreadyRegistered, nil
	default:
		return uuid.Nil, 0, httpx.ErrUnauthorized
	}
}

func (h *Handler) registerFresh(
	w http.ResponseWriter, r *http.Request,
	email, passwordHash string,
	displayName httpx.Optional[string],
) {
	user, err := h.queries.CreateRegisteredUser(r.Context(), db.CreateRegisteredUserParams{
		Email:        pgtype.Text{String: email, Valid: true},
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
		DisplayName:  optionalToPgText(displayName),
	})
	if isUniqueViolation(err) {
		httpx.WriteError(w, httpx.ErrEmailTaken)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "create registered user failed", "error", err)
		httpx.WriteError(w, fmt.Errorf("create registered user: %w", err))
		return
	}

	userID, _ := uuid.FromBytes(user.ID.Bytes[:])
	token, err := h.tokens.Sign(userID, RoleRegistered)
	if err != nil {
		slog.ErrorContext(r.Context(), "sign register token failed", "error", err, "user_id", userID)
		httpx.WriteError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "user registered", "user_id", userID)
	httpx.WriteJSON(w, http.StatusCreated, AuthResponse{
		User:  ToUserResponse(user),
		Token: token,
	})
}

func (h *Handler) registerConvert(
	w http.ResponseWriter, r *http.Request,
	guestID uuid.UUID,
	email, passwordHash string,
	displayName httpx.Optional[string],
) {
	user, err := ConvertGuest(r.Context(), h.pool, h.queries, guestID, email, passwordHash, displayName)
	if err != nil {
		if isUnexpectedError(err) {
			slog.ErrorContext(r.Context(), "convert guest failed", "error", err, "user_id", guestID)
		}
		httpx.WriteError(w, err)
		return
	}

	token, err := h.tokens.Sign(guestID, RoleRegistered)
	if err != nil {
		slog.ErrorContext(r.Context(), "sign convert token failed", "error", err, "user_id", guestID)
		httpx.WriteError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "guest converted", "user_id", guestID)
	httpx.WriteJSON(w, http.StatusCreated, AuthResponse{
		User:  ToUserResponse(user),
		Token: token,
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// isUnexpectedError returns true for errors that should be logged at Error
// level. Expected, client-facing errors (email taken, not found) are not.
func isUnexpectedError(err error) bool {
	switch {
	case errors.Is(err, httpx.ErrEmailTaken),
		errors.Is(err, httpx.ErrNotFound),
		errors.Is(err, httpx.ErrAlreadyRegistered),
		errors.Is(err, httpx.ErrInvalidCredentials),
		errors.Is(err, httpx.ErrUnauthorized):
		return false
	}
	return true
}

