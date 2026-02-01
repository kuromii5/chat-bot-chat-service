package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

var (
	ErrAuthorizationHeaderRequired = errors.New("authorization header is required")
	ErrInvalidAuthorizationFormat  = errors.New("invalid authorization format")
	ErrInvalidOrExpiredToken       = errors.New("invalid or expired token")
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserRoleKey contextKey = "userRole"

func Auth(jwtManager *jwt.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				wrapper.WrapError(w, r, ErrAuthorizationHeaderRequired)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				wrapper.WrapError(w, r, ErrInvalidAuthorizationFormat)
				return
			}

			claims, err := jwtManager.Verify(parts[1])
			if err != nil {
				wrapper.WrapError(w, r, ErrInvalidOrExpiredToken)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
