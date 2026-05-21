package server

import (
	"net/http"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type routeRegistrar interface {
	RegisterRoutes(r chi.Router)
}

type Deps struct {
	AuthMW          func(http.Handler) http.Handler
	AuthHandler     routeRegistrar
	UserHandler     routeRegistrar
	DeviceHandler   routeRegistrar
	EQHandler       routeRegistrar
	PartyHandler    routeRegistrar
	FirmwareHandler routeRegistrar
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mount(r, "/auth", deps.AuthHandler)

	r.Group(func(r chi.Router) {
		if deps.AuthMW != nil {
			r.Use(deps.AuthMW)
		}

		mount(r, "/users", deps.UserHandler)
		mount(r, "/devices", deps.DeviceHandler)
		mount(r, "/eq-profiles", deps.EQHandler)
		mount(r, "/party-sessions", deps.PartyHandler)
		mount(r, "/firmware", deps.FirmwareHandler)
	})

	return r
}

func mount(r chi.Router, pattern string, handler routeRegistrar) {
	if handler == nil {
		return
	}
	r.Route(pattern, handler.RegisterRoutes)
}
