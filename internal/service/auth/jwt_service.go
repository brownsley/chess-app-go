package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateTokenPair(userID, email, name string) (accessToken string, refreshToken string, err error)
	ValidateToken(tokenStr string) (*CustomClaims, error)
}

type jwtService struct {
	secretKey       []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
}

func NewJWTService(secret string, accessDuration, refreshDuration time.Duration) JWTService {
	return &jwtService{
		secretKey:       []byte(secret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

func (s *jwtService) GenerateTokenPair(userID, email, name string) (string, string, error) {
	now := time.Now()

	accessClaims := CustomClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "access",
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := CustomClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "refresh",
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessStr, refreshStr, nil
}

func (s *jwtService) ValidateToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
