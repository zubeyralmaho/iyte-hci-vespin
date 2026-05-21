package auth

import (
	"errors"
	"testing"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if err := VerifyPassword(hash, "correct-horse-battery-staple"); err != nil {
		t.Errorf("VerifyPassword (correct): %v", err)
	}
}

func TestVerifyPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := HashPassword("right")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	err = VerifyPassword(hash, "wrong")
	if err == nil {
		t.Fatal("VerifyPassword: expected error for wrong password, got nil")
	}
	if !errors.Is(err, httpx.ErrInvalidCredentials) {
		t.Errorf("VerifyPassword: error should wrap ErrInvalidCredentials; got %v", err)
	}
}

func TestHashPasswordProducesDistinctSalts(t *testing.T) {
	password := "the-same-password"

	a, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword a: %v", err)
	}
	b, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword b: %v", err)
	}

	if a == b {
		t.Error("two hashes of the same password should differ (bcrypt salts)")
	}
	if err := VerifyPassword(a, password); err != nil {
		t.Errorf("VerifyPassword a: %v", err)
	}
	if err := VerifyPassword(b, password); err != nil {
		t.Errorf("VerifyPassword b: %v", err)
	}
}
