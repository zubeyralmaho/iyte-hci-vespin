package users

import (
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(q *db.Queries) *Handler {
	return &Handler{queries: q}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
}
