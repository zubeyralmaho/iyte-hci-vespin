package partysessions

import (
	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateRequest is the body for POST /party-sessions. Name is optional;
// an empty/omitted name becomes NULL in the database. DeviceIds must
// contain at least one device, and every device must be owned by the
// caller (validated in the service layer).
type CreateRequest struct {
	Name      string      `json:"name"      validate:"omitempty,min=1,max=100"`
	DeviceIds []uuid.UUID `json:"deviceIds" validate:"required,min=1,max=20"`
}

// UpdateRequest is the PATCH body for /party-sessions/{id}. Both fields
// are optional. Empty body is valid and re-writes the row with current
// values (ticks updated_at), consistent with PATCH semantics elsewhere.
type UpdateRequest struct {
	Name   httpx.Optional[string] `json:"name"`
	Status httpx.Optional[string] `json:"status"`
}

// AddDeviceRequest is the body for POST /party-sessions/{id}/devices.
type AddDeviceRequest struct {
	DeviceID uuid.UUID `json:"deviceId" validate:"required"`
}

// Response mirrors the PartySession schema from openapi.yaml.
type Response struct {
	ID        uuid.UUID   `json:"id"`
	Name      *string     `json:"name"`
	Status    string      `json:"status"`
	StartedAt string      `json:"startedAt"`
	EndedAt   *string     `json:"endedAt"`
	DeviceIds []uuid.UUID `json:"deviceIds"`
	CreatedAt string      `json:"createdAt"`
}

// sessionView is the canonical in-package shape for a session plus its
// device-id list. The three sqlc Row types (Get, ListByOwner,
// ListByOwnerAndStatus) all carry the same fields with different concrete
// type names; we collapse them through this view so ToResponse only
// understands one shape.
type sessionView struct {
	ID        pgtype.UUID
	Name      pgtype.Text
	Status    string
	StartedAt pgtype.Timestamptz
	EndedAt   pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeviceIds []pgtype.UUID
}

func viewFromGetRow(r db.GetPartySessionByIDAndOwnerRow) sessionView {
	return sessionView{
		ID:        r.ID,
		Name:      r.Name,
		Status:    r.Status,
		StartedAt: r.StartedAt,
		EndedAt:   r.EndedAt,
		CreatedAt: r.CreatedAt,
		DeviceIds: r.DeviceIds,
	}
}

func viewFromListRow(r db.ListPartySessionsByOwnerRow) sessionView {
	return sessionView{
		ID:        r.ID,
		Name:      r.Name,
		Status:    r.Status,
		StartedAt: r.StartedAt,
		EndedAt:   r.EndedAt,
		CreatedAt: r.CreatedAt,
		DeviceIds: r.DeviceIds,
	}
}

func viewFromListByStatusRow(r db.ListPartySessionsByOwnerAndStatusRow) sessionView {
	return sessionView{
		ID:        r.ID,
		Name:      r.Name,
		Status:    r.Status,
		StartedAt: r.StartedAt,
		EndedAt:   r.EndedAt,
		CreatedAt: r.CreatedAt,
		DeviceIds: r.DeviceIds,
	}
}

func ToResponse(v sessionView) Response {
	resp := Response{
		Status:    v.Status,
		StartedAt: v.StartedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt: v.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
		DeviceIds: make([]uuid.UUID, len(v.DeviceIds)),
	}
	resp.ID, _ = uuid.FromBytes(v.ID.Bytes[:])
	if v.Name.Valid {
		s := v.Name.String
		resp.Name = &s
	}
	if v.EndedAt.Valid {
		s := v.EndedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		resp.EndedAt = &s
	}
	for i, id := range v.DeviceIds {
		resp.DeviceIds[i], _ = uuid.FromBytes(id.Bytes[:])
	}
	return resp
}

func ToListResponseFromOwner(rows []db.ListPartySessionsByOwnerRow) []Response {
	out := make([]Response, len(rows))
	for i, row := range rows {
		out[i] = ToResponse(viewFromListRow(row))
	}
	return out
}

func ToListResponseFromOwnerAndStatus(rows []db.ListPartySessionsByOwnerAndStatusRow) []Response {
	out := make([]Response, len(rows))
	for i, row := range rows {
		out[i] = ToResponse(viewFromListByStatusRow(row))
	}
	return out
}
