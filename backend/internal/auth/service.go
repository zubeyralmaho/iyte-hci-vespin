package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sqlStateUniqueViolation is the Postgres SQLSTATE for a unique-constraint
// violation (Class 23 — integrity constraint violation, specific code 505).
const sqlStateUniqueViolation = "23505"

// ConvertGuest promotes a guest user to registered in a single transaction.
// The user's ID does not change; their devices, EQ profiles, and party
// sessions remain attached. Returns ErrEmailTaken when the email is already
// claimed by another row.
func ConvertGuest(
	ctx context.Context,
	pool *pgxpool.Pool,
	q *db.Queries,
	guestID uuid.UUID,
	email string,
	passwordHash string,
	displayName httpx.Optional[string],
) (db.User, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return db.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := q.WithTx(tx)

	emailParam := pgtype.Text{String: email, Valid: true}

	exists, err := qtx.EmailExists(ctx, emailParam)
	if err != nil {
		return db.User{}, fmt.Errorf("email exists: %w", err)
	}
	if exists {
		return db.User{}, httpx.ErrEmailTaken
	}

	row, err := qtx.ConvertGuestToRegistered(ctx, db.ConvertGuestToRegisteredParams{
		ID:           uuidToPgtype(guestID),
		Email:        emailParam,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
		DisplayName:  optionalToPgText(displayName),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.User{}, httpx.ErrNotFound
	}
	if isUniqueViolation(err) {
		return db.User{}, httpx.ErrEmailTaken
	}
	if err != nil {
		return db.User{}, fmt.Errorf("convert guest: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, fmt.Errorf("commit tx: %w", err)
	}
	return row, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func optionalToPgText(o httpx.Optional[string]) pgtype.Text {
	if !o.Set || o.Null {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: o.Value, Valid: true}
}
