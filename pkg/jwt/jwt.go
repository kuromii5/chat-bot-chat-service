package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
}

func Verify(tokenStr string, jwtSecret string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
