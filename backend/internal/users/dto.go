package users

import (
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
)

// UpdateRequest is the body of PATCH /users/me. Only displayName is editable.
// An omitted field leaves the existing value alone; an empty string clears it.
type UpdateRequest struct {
	DisplayName httpx.Optional[string] `json:"displayName"`
}

// PreferencesResponse mirrors the OpenAPI Preferences schema.
type PreferencesResponse struct {
	Theme                string `json:"theme"`
	Language             string `json:"language"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
}

// PreferencesUpdateRequest carries an optional patch for each preference
// field. Omitted fields are layered over the current row (or defaults if no
// row exists) before upserting.
type PreferencesUpdateRequest struct {
	Theme                httpx.Optional[string] `json:"theme"`
	Language             httpx.Optional[string] `json:"language"`
	NotificationsEnabled httpx.Optional[bool]   `json:"notificationsEnabled"`
}

// Default preference values returned when the user has no preferences row.
const (
	defaultTheme                = "system"
	defaultLanguage             = "en"
	defaultNotificationsEnabled = true
)

func ToPreferencesResponse(p db.UserPreference) PreferencesResponse {
	return PreferencesResponse{
		Theme:                p.Theme,
		Language:             p.Language,
		NotificationsEnabled: p.NotificationsEnabled,
	}
}

func defaultPreferences() PreferencesResponse {
	return PreferencesResponse{
		Theme:                defaultTheme,
		Language:             defaultLanguage,
		NotificationsEnabled: defaultNotificationsEnabled,
	}
}
