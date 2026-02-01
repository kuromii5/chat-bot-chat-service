package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kuromii5/chat-bot-auth-service/pkg/jwt"
)

func NewRouter(chatHandler *Handler, jwtManager *jwt.JWTManager) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	r.Route("/api/v1/chat", func(r chi.Router) {
		r.Use(Auth(jwtManager))

		r.Post("/send", chatHandler.SendMessage)
	})

	return r
}
