// Package auth provides authentication services including password hashing, JWT token management, and Clerk integration.
package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	golangjwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// JWTClaims represents standard and custom JWT claims.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	golangjwt.RegisteredClaims
}

// AuthService encapsulates authentication logic for JWT tokens and password hashing.
type AuthService struct {
	jwtSecret          []byte
	clerkSecretKey     string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	BcryptCost         int
}

// NewAuthService initializes an AuthService instance.
func NewAuthService(jwtSecret string, clerkSecretKey string) *AuthService {
	if clerkSecretKey != "" {
		clerk.SetKey(clerkSecretKey)
	}

	return &AuthService{
		jwtSecret:       []byte(jwtSecret),
		clerkSecretKey:  clerkSecretKey,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		BcryptCost:      12,
	}
}

// HashPassword hashes a plain text password using bcrypt.
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword verifies a plain text password against a bcrypt hash.
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateTokenPair issues a new JWT Access Token and Refresh Token for a user.
func (s *AuthService) GenerateTokenPair(userID, email string) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	// Access Token
	accessClaims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: golangjwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "flux",
			ExpiresAt: golangjwt.NewNumericDate(now.Add(s.AccessTokenTTL)),
			IssuedAt:  golangjwt.NewNumericDate(now),
		},
	}
	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, accessClaims)
	accessToken, err = token.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := &JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: golangjwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    "flux-refresh",
			ExpiresAt: golangjwt.NewNumericDate(now.Add(s.RefreshTokenTTL)),
			IssuedAt:  golangjwt.NewNumericDate(now),
		},
	}
	refreshTokenObj := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a signed JWT Access Token string.
func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	token, err := golangjwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *golangjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken parses and validates a signed JWT Refresh Token string.
func (s *AuthService) ValidateRefreshToken(tokenStr string) (*JWTClaims, error) {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	if claims.Issuer != "flux-refresh" {
		return nil, errors.New("invalid token issuer for refresh token")
	}

	return claims, nil
}

// VerifyClerkToken verifies a Clerk session token using clerk-sdk-go/v2.
func (s *AuthService) VerifyClerkToken(ctx context.Context, sessionToken string) (*clerk.SessionClaims, error) {
	claims, err := clerkjwt.Verify(ctx, &clerkjwt.VerifyParams{
		Token: sessionToken,
	})
	if err != nil {
		return nil, fmt.Errorf("clerk token verification failed: %w", err)
	}
	return claims, nil
}
