package auth_test

import (
	"testing"
	"time"

	"flux/apps/backend/internal/modules/auth"
)

func TestAuthService_PasswordHashing(t *testing.T) {
	svc := auth.NewAuthService("test-secret-key", "")

	password := "SecurePass123!"
	hash, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error hashing password, got: %v", err)
	}

	if hash == "" || hash == password {
		t.Fatalf("expected valid non-empty hash different from raw password")
	}

	if !svc.CheckPassword(password, hash) {
		t.Errorf("expected CheckPassword to return true for matching password")
	}

	if svc.CheckPassword("WrongPassword!", hash) {
		t.Errorf("expected CheckPassword to return false for incorrect password")
	}
}

func TestAuthService_GenerateAndValidateTokenPair(t *testing.T) {
	secret := "test-jwt-secret-key-12345"
	svc := auth.NewAuthService(secret, "")

	userID := "user_123"
	email := "user@example.com"

	accessToken, refreshToken, err := svc.GenerateTokenPair(userID, email)
	if err != nil {
		t.Fatalf("expected no error generating token pair, got: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatalf("expected non-empty access and refresh tokens")
	}

	// Validate Access Token
	claims, err := svc.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("expected valid token validation, got: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected UserID '%s', got '%s'", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("expected Email '%s', got '%s'", email, claims.Email)
	}

	// Validate Refresh Token
	refreshClaims, err := svc.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token validation, got: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("expected refresh UserID '%s', got '%s'", userID, refreshClaims.UserID)
	}
}

func TestAuthService_ValidateToken_InvalidSecret(t *testing.T) {
	svc1 := auth.NewAuthService("secret-one", "")
	svc2 := auth.NewAuthService("secret-two", "")

	token, _, err := svc1.GenerateTokenPair("user_123", "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Errorf("expected error validating token with wrong secret, got nil")
	}
}

func TestAuthService_ValidateToken_Expired(t *testing.T) {
	svc := auth.NewAuthService("test-secret", "")
	svc.AccessTokenTTL = -1 * time.Minute // Expired 1 minute ago

	token, _, err := svc.GenerateTokenPair("user_123", "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = svc.ValidateToken(token)
	if err == nil {
		t.Errorf("expected error validating expired token, got nil")
	}
}
