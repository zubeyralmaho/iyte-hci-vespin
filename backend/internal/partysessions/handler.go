package partysessions

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
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewHandler(q *db.Queries, pool *pgxpool.Pool) *Handler {
	return &Handler{queries: q, pool: pool}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/devices", h.addDevice)
	r.Delete("/{id}/devices/{deviceId}", h.removeDevice)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	statusFilter := r.URL.Query().Get("status")

	if statusFilter == "" {
		rows, err := h.queries.ListPartySessionsByOwner(r.Context(), uuidToPg(userID))
		if err != nil {
			slog.ErrorContext(r.Context(), "list party sessions failed",
				"error", err, "user_id", userID)
			httpx.WriteError(w, fmt.Errorf("list party sessions: %w", err))
			return
		}
		httpx.WriteJSON(w, http.StatusOK, ToListResponseFromOwner(rows))
		return
	}

	if !validStatus(statusFilter) {
		httpx.WriteError(w, httpx.NewValidationError(
			map[string]string{"status": "must be one of: active paused ended"}, nil))
		return
	}

	rows, err := h.queries.ListPartySessionsByOwnerAndStatus(r.Context(), db.ListPartySessionsByOwnerAndStatusParams{
		OwnerUserID: uuidToPg(userID),
		Status:      statusFilter,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "list party sessions by status failed",
			"error", err, "user_id", userID, "status", statusFilter)
		httpx.WriteError(w, fmt.Errorf("list party sessions: %w", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToListResponseFromOwnerAndStatus(rows))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	session, err := Create(r.Context(), h.pool, h.queries, userID, req.Name, req.DeviceIds)
	if err != nil {
		if !isExpectedClientError(err) {
			slog.ErrorContext(r.Context(), "create party session failed",
				"error", err, "user_id", userID)
		}
		httpx.WriteError(w, err)
		return
	}

	sessionID, _ := uuid.FromBytes(session.ID.Bytes[:])
	slog.InfoContext(r.Context(), "party session created",
		"party_session_id", sessionID, "user_id", userID, "device_count", len(session.DeviceIds))
	httpx.WriteJSON(w, http.StatusCreated, ToResponse(viewFromGetRow(session)))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseIDParam(r, "id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	userID := auth.UserIDFromContext(r.Context())

	row, err := h.queries.GetPartySessionByIDAndOwner(r.Context(), db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "get party session failed",
			"error", err, "user_id", userID, "party_session_id", sessionID)
		httpx.WriteError(w, fmt.Errorf("get party session: %w", err))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToResponse(viewFromGetRow(row)))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseIDParam(r, "id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req UpdateRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	row, err := Update(r.Context(), h.pool, h.queries, sessionID, userID, req)
	if err != nil {
		if !isExpectedClientError(err) {
			slog.ErrorContext(r.Context(), "update party session failed",
				"error", err, "user_id", userID, "party_session_id", sessionID)
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToResponse(viewFromGetRow(row)))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseIDParam(r, "id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	userID := auth.UserIDFromContext(r.Context())

	rows, err := h.queries.DeletePartySession(r.Context(), db.DeletePartySessionParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "delete party session failed",
			"error", err, "user_id", userID, "party_session_id", sessionID)
		httpx.WriteError(w, fmt.Errorf("delete party session: %w", err))
		return
	}
	if rows == 0 {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) addDevice(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseIDParam(r, "id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	var req AddDeviceRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	row, err := AddDevice(r.Context(), h.queries, sessionID, req.DeviceID, userID)
	if err != nil {
		if !isExpectedClientError(err) {
			slog.ErrorContext(r.Context(), "add device to party session failed",
				"error", err, "user_id", userID,
				"party_session_id", sessionID, "device_id", req.DeviceID)
		}
		httpx.WriteError(w, err)
		return
	}
	slog.InfoContext(r.Context(), "device added to party session",
		"party_session_id", sessionID, "device_id", req.DeviceID, "user_id", userID)
	httpx.WriteJSON(w, http.StatusCreated, ToResponse(viewFromGetRow(row)))
}

func (h *Handler) removeDevice(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseIDParam(r, "id")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	deviceID, err := parseIDParam(r, "deviceId")
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	row, err := RemoveDevice(r.Context(), h.queries, sessionID, deviceID, userID)
	if err != nil {
		if !isExpectedClientError(err) {
			slog.ErrorContext(r.Context(), "remove device from party session failed",
				"error", err, "user_id", userID,
				"party_session_id", sessionID, "device_id", deviceID)
		}
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, ToResponse(viewFromGetRow(row)))
}

// isExpectedClientError reports whether err is one of the domain errors
// that maps to a 4xx response. We skip the error-level log line for those
// since they are routine client mistakes, not server faults.
func isExpectedClientError(err error) bool {
	var ve *httpx.ValidationError
	if errors.As(err, &ve) {
		return true
	}
	switch {
	case errors.Is(err, httpx.ErrNotFound),
		errors.Is(err, httpx.ErrInvalidDeviceRef),
		errors.Is(err, httpx.ErrInvalidStatusTransition),
		errors.Is(err, httpx.ErrDeviceAlreadyInSession):
		return true
	}
	return false
}

func parseIDParam(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, httpx.NewValidationError(
			map[string]string{name: "must be a valid UUID"}, err)
	}
	return id, nil
}
