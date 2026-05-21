package partysessions

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sqlStateUniqueViolation is the Postgres SQLSTATE for a unique-constraint
// violation. Used to convert PK conflicts on party_session_devices to the
// domain-specific ErrDeviceAlreadyInSession.
const sqlStateUniqueViolation = "23505"

// Create transactionally validates that every supplied device belongs to
// userID, inserts a new party session, and adds the devices as members.
// The session is returned in its post-insert form (re-fetched outside the
// tx so the array_agg of device_ids is populated).
func Create(
	ctx context.Context,
	pool *pgxpool.Pool,
	q *db.Queries,
	userID uuid.UUID,
	name string,
	deviceIDs []uuid.UUID,
) (db.GetPartySessionByIDAndOwnerRow, error) {
	uniqueIDs := dedupeUUIDs(deviceIDs)
	pgIDs := make([]pgtype.UUID, len(uniqueIDs))
	for i, id := range uniqueIDs {
		pgIDs[i] = uuidToPg(id)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := q.WithTx(tx)

	count, err := qtx.CountDevicesByIDsAndUser(ctx, db.CountDevicesByIDsAndUserParams{
		UserID:    uuidToPg(userID),
		DeviceIds: pgIDs,
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("count devices: %w", err)
	}
	if int(count) != len(uniqueIDs) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrInvalidDeviceRef
	}

	created, err := qtx.CreatePartySession(ctx, db.CreatePartySessionParams{
		OwnerUserID: uuidToPg(userID),
		Name:        textFromString(name),
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("create party session: %w", err)
	}

	if _, err := qtx.AddPartySessionDevices(ctx, db.AddPartySessionDevicesParams{
		PartySessionID: created.ID,
		OwnerUserID:    uuidToPg(userID),
		DeviceIds:      pgIDs,
	}); err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("add party session devices: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("commit tx: %w", err)
	}

	full, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          created.ID,
		OwnerUserID: uuidToPg(userID),
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("refetch party session: %w", err)
	}
	return full, nil
}

// Update applies a PATCH to a party session in a single transaction:
// SELECT current → validate status transition → UPDATE. The full row
// (with device_ids) is re-fetched after commit. Omitted Optional fields
// preserve their current value.
func Update(
	ctx context.Context,
	pool *pgxpool.Pool,
	q *db.Queries,
	sessionID, userID uuid.UUID,
	req UpdateRequest,
) (db.GetPartySessionByIDAndOwnerRow, error) {
	if req.Status.Set && !req.Status.Null && !validStatus(req.Status.Value) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.NewValidationError(
			map[string]string{"status": "must be one of: active paused ended"}, nil)
	}
	if req.Name.Set && !req.Name.Null {
		trimmed := strings.TrimSpace(req.Name.Value)
		if n := len(trimmed); n < 1 || n > 100 {
			return db.GetPartySessionByIDAndOwnerRow{}, httpx.NewValidationError(
				map[string]string{"name": "must be between 1 and 100 characters"}, nil)
		}
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := q.WithTx(tx)

	current, err := qtx.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrNotFound
	}
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("get party session: %w", err)
	}

	name := current.Name
	if req.Name.Set && !req.Name.Null {
		name = pgtype.Text{String: strings.TrimSpace(req.Name.Value), Valid: true}
	}

	status := current.Status
	if req.Status.Set && !req.Status.Null {
		if !legalTransition(current.Status, req.Status.Value) {
			return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrInvalidStatusTransition
		}
		status = req.Status.Value
	}

	if _, err := qtx.UpdatePartySession(ctx, db.UpdatePartySessionParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
		Name:        name,
		Status:      status,
	}); errors.Is(err, pgx.ErrNoRows) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrNotFound
	} else if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("update party session: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("commit tx: %w", err)
	}

	full, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("refetch party session: %w", err)
	}
	return full, nil
}

// AddDevice validates the session exists, is not ended, and that deviceID
// is owned by the caller before inserting a membership row. A PK conflict
// is mapped to ErrDeviceAlreadyInSession.
func AddDevice(
	ctx context.Context,
	q *db.Queries,
	sessionID, deviceID, userID uuid.UUID,
) (db.GetPartySessionByIDAndOwnerRow, error) {
	session, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrNotFound
	}
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("get party session: %w", err)
	}
	if session.Status == "ended" {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrInvalidStatusTransition
	}

	if _, err := q.GetDeviceByIDAndUser(ctx, db.GetDeviceByIDAndUserParams{
		ID:     uuidToPg(deviceID),
		UserID: uuidToPg(userID),
	}); errors.Is(err, pgx.ErrNoRows) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrInvalidDeviceRef
	} else if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("get device: %w", err)
	}

	rows, err := q.AddPartySessionDevice(ctx, db.AddPartySessionDeviceParams{
		PartySessionID: uuidToPg(sessionID),
		DeviceID:       uuidToPg(deviceID),
		OwnerUserID:    uuidToPg(userID),
	})
	if isUniqueViolation(err) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrDeviceAlreadyInSession
	}
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("add party session device: %w", err)
	}
	if rows == 0 {
		// EXISTS check inside the query failed — the session disappeared
		// between our SELECT and the INSERT.
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrNotFound
	}

	full, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("refetch party session: %w", err)
	}
	return full, nil
}

// RemoveDevice deletes a single membership row. Returns the refreshed
// session. Removing from an ended session is rejected because ended
// sessions are frozen.
func RemoveDevice(
	ctx context.Context,
	q *db.Queries,
	sessionID, deviceID, userID uuid.UUID,
) (db.GetPartySessionByIDAndOwnerRow, error) {
	session, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrNotFound
	}
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("get party session: %w", err)
	}
	if session.Status == "ended" {
		return db.GetPartySessionByIDAndOwnerRow{}, httpx.ErrInvalidStatusTransition
	}

	if _, err := q.RemovePartySessionDevice(ctx, db.RemovePartySessionDeviceParams{
		PartySessionID: uuidToPg(sessionID),
		DeviceID:       uuidToPg(deviceID),
		OwnerUserID:    uuidToPg(userID),
	}); err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("remove party session device: %w", err)
	}
	// A no-op delete (device wasn't in the session) is treated as success;
	// the openapi spec is silent and the resulting state matches the
	// caller's intent. The refetch reflects the actual membership.

	full, err := q.GetPartySessionByIDAndOwner(ctx, db.GetPartySessionByIDAndOwnerParams{
		ID:          uuidToPg(sessionID),
		OwnerUserID: uuidToPg(userID),
	})
	if err != nil {
		return db.GetPartySessionByIDAndOwnerRow{}, fmt.Errorf("refetch party session: %w", err)
	}
	return full, nil
}

func uuidToPg(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func textFromString(s string) pgtype.Text {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func dedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation
}
