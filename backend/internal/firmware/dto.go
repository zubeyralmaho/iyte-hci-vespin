package firmware

import "github.com/google/uuid"

// CheckRequest is the body for POST /firmware/check.
type CheckRequest struct {
	DeviceID       uuid.UUID `json:"deviceId"       validate:"required"`
	CurrentVersion string    `json:"currentVersion" validate:"required,min=1,max=32"`
}

// CheckResponse mirrors the FirmwareCheckResponse schema in openapi.yaml.
type CheckResponse struct {
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseNotes    string `json:"releaseNotes"`
}
