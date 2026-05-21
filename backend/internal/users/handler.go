package users

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/auth"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{queries: q}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/me", h.getMe)
	r.Patch("/me", h.patchMe)

	r.Route("/me/preferences", func(r chi.Router) {
		r.Use(auth.RequireRegistered)
		r.Get("/", h.getPreferences)
		r.Patch("/", h.patchPreferences)
	})
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	user, err := h.queries.GetUserByID(r.Context(), uuidToPg(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "get user failed", "error", err, "user_id", userID)
		httpx.WriteError(w, fmt.Errorf("get user: %w", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, auth.ToUserResponse(user))
}

func (h *Handler) patchMe(w http.ResponseWriter, r *http.Request) {
	var req UpdateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())

	if !req.DisplayName.Set {
		// Empty body or only-other-fields: return current state unchanged.
		user, err := h.queries.GetUserByID(r.Context(), uuidToPg(userID))
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, httpx.ErrNotFound)
			return
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "get user failed", "error", err, "user_id", userID)
			httpx.WriteError(w, fmt.Errorf("get user: %w", err))
			return
		}
		httpx.WriteJSON(w, http.StatusOK, auth.ToUserResponse(user))
		return
	}

	dn := pgtype.Text{Valid: false}
	if !req.DisplayName.Null {
		if len(req.DisplayName.Value) > 100 {
			httpx.WriteError(w, httpx.NewValidationError(
				map[string]string{"displayName": "must be at most 100"}, nil))
			return
		}
		dn = pgtype.Text{String: req.DisplayName.Value, Valid: true}
	}

	user, err := h.queries.UpdateUserDisplayName(r.Context(), db.UpdateUserDisplayNameParams{
		ID:          uuidToPg(userID),
		DisplayName: dn,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "update display name failed", "error", err, "user_id", userID)
		httpx.WriteError(w, fmt.Errorf("update display name: %w", err))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, auth.ToUserResponse(user))
}

func (h *Handler) getPreferences(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	row, err := h.queries.GetUserPreferences(r.Context(), uuidToPg(userID))
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteJSON(w, http.StatusOK, defaultPreferences())
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "get preferences failed", "error", err, "user_id", userID)
		httpx.WriteError(w, fmt.Errorf("get preferences: %w", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToPreferencesResponse(row))
}

func (h *Handler) patchPreferences(w http.ResponseWriter, r *http.Request) {
	var req PreferencesUpdateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	merged, err := mergePreferences(r, req, h)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	row, err := h.queries.UpsertUserPreferences(r.Context(), db.UpsertUserPreferencesParams{
		UserID:               uuidToPg(userID),
		Theme:                merged.Theme,
		Language:             merged.Language,
		NotificationsEnabled: merged.NotificationsEnabled,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "upsert preferences failed", "error", err, "user_id", userID)
		httpx.WriteError(w, fmt.Errorf("upsert preferences: %w", err))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, ToPreferencesResponse(row))
}

// mergePreferences validates the patch and layers Set fields over the user's
// current preferences (or defaults if no row exists yet).
func mergePreferences(r *http.Request, req PreferencesUpdateRequest, h *Handler) (PreferencesResponse, error) {
	userID := auth.UserIDFromContext(r.Context())

	current := defaultPreferences()
	row, err := h.queries.GetUserPreferences(r.Context(), uuidToPg(userID))
	if err == nil {
		current = ToPreferencesResponse(row)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return PreferencesResponse{}, fmt.Errorf("get preferences: %w", err)
	}

	fields := map[string]string{}

	if req.Theme.Set && !req.Theme.Null {
		switch req.Theme.Value {
		case "light", "dark", "system":
			current.Theme = req.Theme.Value
		default:
			fields["theme"] = "must be one of: light dark system"
		}
	}

	if req.Language.Set && !req.Language.Null {
		switch n := len(req.Language.Value); {
		case n < 2:
			fields["language"] = "must be at least 2"
		case n > 8:
			fields["language"] = "must be at most 8"
		default:
			current.Language = req.Language.Value
		}
	}

	if req.NotificationsEnabled.Set && !req.NotificationsEnabled.Null {
		current.NotificationsEnabled = req.NotificationsEnabled.Value
	}

	if len(fields) > 0 {
		return PreferencesResponse{}, httpx.NewValidationError(fields, nil)
	}

	return current, nil
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}
