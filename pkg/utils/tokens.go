package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MyClaims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uuid.UUID, username string, secret string, tokenType string) (string, error) {
	var exp time.Duration
	if tokenType == "access" {
		exp = time.Hour * 24
	} else if tokenType == "refresh" {
		exp = time.Hour * 24 * 7
	} else {
		exp = time.Hour * 1 // Default 1 hour
	}

	claims := &MyClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyToken(tokenString string, secret string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
