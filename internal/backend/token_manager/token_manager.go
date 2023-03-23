//go:generate mockgen -destination=../../mocks/token_manager.go -package=mocks github.com/apolsh/yapr-gophkeeper/internal/backend/token_manager TokenManager
package token_manager

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenManager interface {
	GenerateToken(id int64) (string, error)
	ParseToken(tokenString string) (int64, error)
}

type jwtTokenClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"user_id"`
}

// JWTTokenManager token manager jwt implementation
type JWTTokenManager struct {
	jwtSecretKey string
}

// NewJWTTokenManager JWTTokenManager constructor
func NewJWTTokenManager(secretKey string) *JWTTokenManager {
	return &JWTTokenManager{jwtSecretKey: secretKey}
}

// GenerateToken generates new token
func (s *JWTTokenManager) GenerateToken(id int64) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtTokenClaims{
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		id,
	})

	return token.SignedString([]byte(s.jwtSecretKey))
}

// ParseToken parses generated token
func (s *JWTTokenManager) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtSecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(*jwtTokenClaims)
	if !ok {
		return 0, errors.New("invalid token claims type")
	}
	return claims.UserID, nil
}
