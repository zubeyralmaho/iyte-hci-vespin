package auth

import (
	"net/http"
	"strings"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
)

func Middleware(tokens *Tokens) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := strings.TrimSpace(r.Header.Get("Authorization"))
			if header == "" {
				httpx.WriteError(w, httpx.ErrUnauthorized)
				return
			}

			tokenValue, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(tokenValue) == "" {
				httpx.WriteError(w, httpx.ErrUnauthorized)
				return
			}

			claims, err := tokens.Verify(strings.TrimSpace(tokenValue))
			if err != nil {
				httpx.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(withUser(r.Context(), claims.UserID, claims.Role)))
		})
	}
}

// RequireRegistered rejects requests whose JWT role is not "registered".
// Must run after Middleware so the role is already in context.
func RequireRegistered(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RoleFromContext(r.Context()) != RoleRegistered {
			httpx.WriteError(w, httpx.ErrGuestEndpointForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
