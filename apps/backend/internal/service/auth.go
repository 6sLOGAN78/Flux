package service

import (
	"context"
	"time"

	clerkjwt "github.com/clerk/clerk-sdk-go/v2/jwt"
	golangjwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtSecret      []byte
	clerkSecretKey string
}

func NewAuthService(jwtSecret string, clerkSecretKey string) *AuthService {
	return &AuthService{
		jwtSecret:      []byte(jwtSecret),
		clerkSecretKey: clerkSecretKey,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func (s *AuthService) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) ValidateToken(tokenString string) (golangjwt.MapClaims, error) {
	// First attempt Clerk JWT session validation if Clerk Secret Key is configured
	if s.clerkSecretKey != "" {
		sessionClaims, err := clerkjwt.Verify(context.Background(), &clerkjwt.VerifyParams{
			Token: tokenString,
		})
		if err == nil && sessionClaims != nil {
			return golangjwt.MapClaims{
				"sub":    sessionClaims.Subject,
				"claims": sessionClaims,
			}, nil
		}
	}

	// Fallback to native HMAC SHA256 JWT verification
	token, err := golangjwt.Parse(tokenString, func(t *golangjwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*golangjwt.SigningMethodHMAC); !ok {
			return nil, golangjwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(golangjwt.MapClaims)
	if !ok {
		return nil, golangjwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func (s *AuthService) GenerateTokenPair(userID string, email string) (accessToken string, refreshToken string, err error) {
	now := time.Now()

	accessClaims := golangjwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   now.Add(15 * time.Minute).Unix(),
		"iat":   now.Unix(),
	}

	accessTokenObj := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	refreshClaims := golangjwt.MapClaims{
		"sub": userID,
		"exp": now.Add(7 * 24 * time.Hour).Unix(),
		"iat": now.Unix(),
	}

	refreshTokenObj := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(s.jwtSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
