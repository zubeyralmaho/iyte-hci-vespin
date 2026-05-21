package firmware

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
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireRegistered)
		r.Post("/check", h.check)
	})
}

func (h *Handler) check(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, err)
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	device, err := h.queries.GetDeviceByIDAndUser(r.Context(), db.GetDeviceByIDAndUserParams{
		ID:     pgtype.UUID{Bytes: req.DeviceID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.WriteError(w, httpx.ErrNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "firmware check get device failed",
			"error", err, "user_id", userID, "device_id", req.DeviceID)
		httpx.WriteError(w, fmt.Errorf("get device: %w", err))
		return
	}

	latestVersion, ok := latest[device.DeviceType]
	if !ok {
		// Closed enum on devices.device_type — this is reachable only if
		// the schema and the versions map drift.
		deviceID, _ := uuid.FromBytes(device.ID.Bytes[:])
		slog.WarnContext(r.Context(), "firmware check missing version map entry",
			"device_id", deviceID, "device_type", device.DeviceType)
		httpx.WriteError(w, fmt.Errorf("no firmware version known for device type %q", device.DeviceType))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, CheckResponse{
		LatestVersion:   latestVersion,
		UpdateAvailable: req.CurrentVersion != latestVersion,
		ReleaseNotes:    "",
	})
}
