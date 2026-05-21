package auth

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/db"
	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConvertGuest tests need a real Postgres. Set TEST_DATABASE_URL to a clean
// (or reusable) database and run:
//
//	TEST_DATABASE_URL=postgres://... go test ./internal/auth/...
//
// Without TEST_DATABASE_URL, these tests are skipped. The tests bring the
// schema up themselves by executing every *.up.sql file in
// internal/db/migrations in order — make sure the target database is empty,
// or willing to skip already-applied migrations (idempotency is not
// guaranteed across runs; use a throwaway DB).
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := runMigrations(ctx, pool); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	dir, err := findMigrationsDir()
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)
	for _, name := range ups {
		path := filepath.Join(dir, name)
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func findMigrationsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// Walk up looking for the project's migrations directory. The auth
	// package sits at backend/internal/auth, so migrations are two levels up
	// in internal/db/migrations.
	dir := wd
	for range 6 {
		candidate := filepath.Join(dir, "internal", "db", "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.New("could not locate internal/db/migrations from " + wd)
}

// uniqueEmail returns a per-run email so concurrent or repeated test
// invocations don't collide on the unique index.
func uniqueEmail(prefix string) string {
	return prefix + "+" + uuid.NewString() + "@example.test"
}

func makeGuest(t *testing.T, q *db.Queries) db.User {
	t.Helper()
	guest, err := q.CreateGuestUser(context.Background())
	if err != nil {
		t.Fatalf("CreateGuestUser: %v", err)
	}
	return guest
}

func TestConvertGuestHappyPath(t *testing.T) {
	pool := openTestPool(t)
	q := db.New(pool)
	ctx := context.Background()

	guest := makeGuest(t, q)
	guestID, _ := uuid.FromBytes(guest.ID.Bytes[:])
	email := uniqueEmail("happy")

	user, err := ConvertGuest(ctx, pool, q, guestID, email, "hash:bcrypt", httpx.Optional[string]{
		Set: true, Value: "Alice",
	})
	if err != nil {
		t.Fatalf("ConvertGuest: %v", err)
	}

	if user.Role != RoleRegistered {
		t.Errorf("Role: got %q, want %q", user.Role, RoleRegistered)
	}
	if !user.Email.Valid || user.Email.String != email {
		t.Errorf("Email: got %+v, want %q", user.Email, email)
	}
	if !user.DisplayName.Valid || user.DisplayName.String != "Alice" {
		t.Errorf("DisplayName: got %+v, want %q", user.DisplayName, "Alice")
	}
	if !user.ConvertedAt.Valid {
		t.Error("ConvertedAt: expected non-null after conversion")
	}

	// User ID must be unchanged (the same row was updated, not a new one).
	gotID, _ := uuid.FromBytes(user.ID.Bytes[:])
	if gotID != guestID {
		t.Errorf("ID changed: got %v, want %v", gotID, guestID)
	}
}

func TestConvertGuestEmailTakenLeavesGuestUntouched(t *testing.T) {
	pool := openTestPool(t)
	q := db.New(pool)
	ctx := context.Background()

	// Create the registered user that owns the email.
	takenEmail := uniqueEmail("taken")
	_, err := q.CreateRegisteredUser(ctx, db.CreateRegisteredUserParams{
		Email:        pgtype.Text{String: takenEmail, Valid: true},
		PasswordHash: pgtype.Text{String: "hash", Valid: true},
		DisplayName:  pgtype.Text{Valid: false},
	})
	if err != nil {
		t.Fatalf("CreateRegisteredUser: %v", err)
	}

	// Create a guest and try to convert with the taken email.
	guest := makeGuest(t, q)
	guestID, _ := uuid.FromBytes(guest.ID.Bytes[:])

	_, err = ConvertGuest(ctx, pool, q, guestID, takenEmail, "hash", httpx.Optional[string]{})
	if !errors.Is(err, httpx.ErrEmailTaken) {
		t.Fatalf("ConvertGuest: expected ErrEmailTaken, got %v", err)
	}

	// Verify the guest row is unchanged.
	row, err := q.GetUserByID(ctx, guest.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if row.Role != RoleGuest {
		t.Errorf("Role: got %q, want %q (guest should be untouched)", row.Role, RoleGuest)
	}
	if row.Email.Valid {
		t.Errorf("Email: got %+v, want null", row.Email)
	}
	if row.ConvertedAt.Valid {
		t.Error("ConvertedAt: got non-null, want null")
	}

	// And the user can still convert with a different email afterwards.
	if _, err := ConvertGuest(ctx, pool, q, guestID, uniqueEmail("retry"), "hash", httpx.Optional[string]{}); err != nil {
		t.Errorf("retry ConvertGuest with fresh email: %v", err)
	}
}

func TestConvertGuestOnRegisteredReturnsNotFound(t *testing.T) {
	pool := openTestPool(t)
	q := db.New(pool)
	ctx := context.Background()

	// A registered user — guest conversion should NOT work on this row.
	regUser, err := q.CreateRegisteredUser(ctx, db.CreateRegisteredUserParams{
		Email:        pgtype.Text{String: uniqueEmail("registered"), Valid: true},
		PasswordHash: pgtype.Text{String: "hash", Valid: true},
		DisplayName:  pgtype.Text{Valid: false},
	})
	if err != nil {
		t.Fatalf("CreateRegisteredUser: %v", err)
	}
	regID, _ := uuid.FromBytes(regUser.ID.Bytes[:])

	_, err = ConvertGuest(ctx, pool, q, regID, uniqueEmail("attempt"), "hash", httpx.Optional[string]{})
	if !errors.Is(err, httpx.ErrNotFound) {
		t.Fatalf("ConvertGuest on registered: expected ErrNotFound, got %v", err)
	}
}

func TestConvertGuestPreservesDisplayNameWhenOmitted(t *testing.T) {
	pool := openTestPool(t)
	q := db.New(pool)
	ctx := context.Background()

	// Guest sets a display name first.
	guest := makeGuest(t, q)
	_, err := q.UpdateUserDisplayName(ctx, db.UpdateUserDisplayNameParams{
		ID:          guest.ID,
		DisplayName: pgtype.Text{String: "Pre-set Name", Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateUserDisplayName: %v", err)
	}

	guestID, _ := uuid.FromBytes(guest.ID.Bytes[:])

	// Convert WITHOUT a displayName field — should preserve the existing one.
	user, err := ConvertGuest(ctx, pool, q, guestID, uniqueEmail("preserve"), "hash", httpx.Optional[string]{})
	if err != nil {
		t.Fatalf("ConvertGuest: %v", err)
	}
	if !user.DisplayName.Valid || user.DisplayName.String != "Pre-set Name" {
		t.Errorf("DisplayName: got %+v, want preserved %q", user.DisplayName, "Pre-set Name")
	}
}
