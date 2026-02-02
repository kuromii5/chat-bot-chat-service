package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(chatHandler *Handler, jwtSecret string) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
		middleware.Timeout(30*time.Second),
	)

	r.Route("/api/v1/chat", func(r chi.Router) {
		r.Use(Auth(jwtSecret))

		r.Route("/profile", func(r chi.Router) {
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", chatHandler.GetProfileTags)
				r.Put("/", chatHandler.UpdateProfileTags)
			})
		})

		r.Post("/send", chatHandler.SendMessage)
	})

	return r
}
