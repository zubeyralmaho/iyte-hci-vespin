package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/ffk00/iyte-hci-vespin/backend/internal/httpx"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestTokensSignAndVerifyRoundtrip(t *testing.T) {
	tokens := NewTokens("test-secret")
	userID := uuid.New()

	signed, err := tokens.Sign(userID, RoleRegistered)
	if err != nil {
		t.Fatalf("Sign: unexpected error: %v", err)
	}

	claims, err := tokens.Verify(signed)
	if err != nil {
		t.Fatalf("Verify: unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID: got %v, want %v", claims.UserID, userID)
	}
	if claims.Role != RoleRegistered {
		t.Errorf("Role: got %q, want %q", claims.Role, RoleRegistered)
	}
}

func TestTokensVerifyRejectsWrongSecret(t *testing.T) {
	signer := NewTokens("secret-a")
	verifier := NewTokens("secret-b")

	signed, err := signer.Sign(uuid.New(), RoleGuest)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	_, err = verifier.Verify(signed)
	if err == nil {
		t.Fatal("Verify: expected error for mismatched secret, got nil")
	}
	if !errors.Is(err, httpx.ErrUnauthorized) {
		t.Errorf("Verify: error should wrap ErrUnauthorized; got %v", err)
	}
}

func TestTokensVerifyRejectsExpired(t *testing.T) {
	tokens := NewTokens("test-secret")

	// Hand-roll an expired token rather than waiting tokenTTL.
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": RoleRegistered,
		"iat":  time.Now().Add(-2 * time.Hour).Unix(),
		"exp":  time.Now().Add(-1 * time.Hour).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tokens.secret)
	if err != nil {
		t.Fatalf("manual sign: %v", err)
	}

	_, err = tokens.Verify(signed)
	if err == nil {
		t.Fatal("Verify: expected error for expired token, got nil")
	}
	if !errors.Is(err, httpx.ErrUnauthorized) {
		t.Errorf("Verify: error should wrap ErrUnauthorized; got %v", err)
	}
}

func TestTokensVerifyRejectsWrongAlgorithm(t *testing.T) {
	tokens := NewTokens("test-secret")

	// "none" alg tokens must be rejected outright.
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": RoleRegistered,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("manual none-sign: %v", err)
	}

	_, err = tokens.Verify(signed)
	if err == nil {
		t.Fatal("Verify: expected error for 'none' alg, got nil")
	}
}

func TestTokensSignRejectsUnknownRole(t *testing.T) {
	tokens := NewTokens("test-secret")

	_, err := tokens.Sign(uuid.New(), "admin")
	if err == nil {
		t.Fatal("Sign: expected error for unknown role, got nil")
	}
}

func TestTokensVerifyRejectsInvalidRoleClaim(t *testing.T) {
	tokens := NewTokens("test-secret")

	// Forge a token with role="admin" (not a valid role) but valid signature.
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": "admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tokens.secret)
	if err != nil {
		t.Fatalf("manual sign: %v", err)
	}

	_, err = tokens.Verify(signed)
	if err == nil {
		t.Fatal("Verify: expected error for unknown role, got nil")
	}
	if !errors.Is(err, httpx.ErrUnauthorized) {
		t.Errorf("Verify: error should wrap ErrUnauthorized; got %v", err)
	}
}
