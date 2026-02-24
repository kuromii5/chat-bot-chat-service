package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"
)

func NewRouter(
	msgHandler *MessageHandler,
	tagHandler *TagHandler,
	notificationHandler *NotificationHandler,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		otelchi.Middleware("chat-service", otelchi.WithChiRoutes(r)),
		middleware.RealIP,
		middleware.Logger,
		middleware.Recoverer,
	)

	r.Group(func(r chi.Router) {
		r.Route("/api/v1/chat", func(r chi.Router) {
			r.Use(Auth(jwtSecret))
			r.Get("/ws", notificationHandler.HandleWS)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Timeout(30 * time.Second))
				r.Route("/profile", func(r chi.Router) {
					r.Route("/tags", func(r chi.Router) {
						r.Get("/", tagHandler.GetProfileTags)
						r.Put("/", tagHandler.UpdateProfileTags)
					})
				})

				r.Post("/send", msgHandler.SendMessage)
			})
		})
	})

	return r
}
