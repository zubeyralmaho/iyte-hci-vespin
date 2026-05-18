package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMapErrorStatusAndCode(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "validation", err: ErrValidationFailed, status: http.StatusBadRequest, code: "validation_failed"},
		{name: "validation details", err: NewValidationError(map[string]string{"email": "is invalid"}, nil), status: http.StatusBadRequest, code: "validation_failed"},
		{name: "unauthorized", err: ErrUnauthorized, status: http.StatusUnauthorized, code: "unauthorized"},
		{name: "forbidden", err: ErrForbidden, status: http.StatusForbidden, code: "forbidden"},
		{name: "guest forbidden", err: ErrGuestEndpointForbidden, status: http.StatusForbidden, code: "guest_endpoint_forbidden"},
		{name: "system preset immutable", err: ErrSystemPresetImmutable, status: http.StatusForbidden, code: "system_preset_immutable"},
		{name: "not found", err: ErrNotFound, status: http.StatusNotFound, code: "not_found"},
		{name: "pgx no rows", err: pgx.ErrNoRows, status: http.StatusNotFound, code: "not_found"},
		{name: "invalid credentials", err: ErrInvalidCredentials, status: http.StatusUnauthorized, code: "invalid_credentials"},
		{name: "email taken", err: ErrEmailTaken, status: http.StatusConflict, code: "email_taken"},
		{name: "already registered", err: ErrAlreadyRegistered, status: http.StatusConflict, code: "already_registered"},
		{name: "conflict", err: ErrConflict, status: http.StatusConflict, code: "conflict"},
		{name: "invalid eq profile", err: ErrInvalidEQProfileRef, status: http.StatusBadRequest, code: "invalid_eq_profile_reference"},
		{name: "not system preset", err: ErrNotASystemPreset, status: http.StatusBadRequest, code: "not_a_system_preset"},
		{name: "invalid device", err: ErrInvalidDeviceRef, status: http.StatusBadRequest, code: "invalid_device_reference"},
		{name: "invalid status transition", err: ErrInvalidStatusTransition, status: http.StatusBadRequest, code: "invalid_status_transition"},
		{name: "device already in session", err: ErrDeviceAlreadyInSession, status: http.StatusConflict, code: "device_already_in_session"},
		{name: "wrapped", err: errors.New("outer: " + ErrConflict.Error()), status: http.StatusInternalServerError, code: "internal_error"},
		{name: "wrapped sentinel", err: fmt.Errorf("context: %w", ErrConflict), status: http.StatusConflict, code: "conflict"},
		{name: "unknown", err: errors.New("boom"), status: http.StatusInternalServerError, code: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _, _ := mapError(tt.err)
			if status != tt.status {
				t.Fatalf("status = %d, want %d", status, tt.status)
			}
			if code != tt.code {
				t.Fatalf("code = %q, want %q", code, tt.code)
			}
		})
	}
}
