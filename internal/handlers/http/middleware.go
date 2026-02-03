package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/pkg/jwt"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserRoleKey contextKey = "userRole"

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				wrapper.WrapError(w, r, domain.ErrAuthorizationHeaderRequired)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				wrapper.WrapError(w, r, domain.ErrInvalidAuthorizationFormat)
				return
			}

			claims, err := jwt.Verify(parts[1], jwtSecret)
			if err != nil {
				wrapper.WrapError(w, r, domain.ErrInvalidOrExpiredToken)
				return
			}

			role := domain.Role(claims.Role)

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
