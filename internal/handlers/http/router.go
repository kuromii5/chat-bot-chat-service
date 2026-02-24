package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	httpMiddleware "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
)

type MessageHandler interface {
	SendMessage(http.ResponseWriter, *http.Request)
}

type TagHandler interface {
	GetProfileTags(http.ResponseWriter, *http.Request)
	UpdateProfileTags(http.ResponseWriter, *http.Request)
}

type WSHandler interface {
	HandleWS(http.ResponseWriter, *http.Request)
}

func NewRouter(
	msgH MessageHandler,
	tagH TagHandler,
	wsH WSHandler,
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
			r.Use(httpMiddleware.Auth(jwtSecret))
			r.Get("/ws", wsH.HandleWS)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Timeout(30 * time.Second))
				r.Route("/profile/tags", func(r chi.Router) {
					r.Get("/", tagH.GetProfileTags)
					r.Put("/", tagH.UpdateProfileTags)
				})
				r.Post("/send", msgH.SendMessage)
			})
		})
	})

	return r
}
