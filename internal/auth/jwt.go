package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type Claims struct {
	UserID    int64     `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) GenerateAccessToken(userID int64) (string, error) {
	return m.generateToken(userID, AccessTokenType, m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID int64) (string, error) {
	return m.generateToken(userID, RefreshTokenType, m.refreshTTL)
}

func (m *Manager) generateToken(userID int64, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (m *Manager) ValidateAccessToken(tokenStr string) (int64, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	if claims.TokenType != AccessTokenType {
		return 0, fmt.Errorf("invalid token type")
	}
	return claims.UserID, nil
}

func (m *Manager) ValidateRefreshToken(tokenStr string) (int64, error) {
	claims, err := m.ParseToken(tokenStr)
	if err != nil {
		return 0, err
	}
	if claims.TokenType != RefreshTokenType {
		return 0, fmt.Errorf("invalid token type")
	}
	return claims.UserID, nil
}