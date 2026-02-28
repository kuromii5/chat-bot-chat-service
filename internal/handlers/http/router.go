package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/riandyrn/otelchi"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	httpMiddleware "github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
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

type RoomHandler interface {
	ClaimRoom(http.ResponseWriter, *http.Request)
	CloseRoom(http.ResponseWriter, *http.Request)
}

func NewRouter(
	msgH MessageHandler,
	tagH TagHandler,
	wsH WSHandler,
	roomH RoomHandler,
	jwtSecret string,
) http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.RequestID,
		otelchi.Middleware("chat-service", otelchi.WithChiRoutes(r)),
		middleware.RealIP,
		wrapper.AccessLog,
		middleware.Recoverer,
	)

	r.Route("/api/v1/chat", func(r chi.Router) {
		r.Use(httpMiddleware.Auth(jwtSecret))

		// Long-lived — no timeout
		r.Get("/ws", wsH.HandleWS)

		// Regular HTTP endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(30 * time.Second))

			// Both roles
			r.Post("/send", msgH.SendMessage)
			r.Post("/rooms/{roomID}/close", roomH.CloseRoom)

			// AI only
			r.Group(func(r chi.Router) {
				r.Use(httpMiddleware.RequireRole(domain.AI))
				r.Get("/profile/tags", tagH.GetProfileTags)
				r.Put("/profile/tags", tagH.UpdateProfileTags)
				r.Post("/rooms/{roomID}/claim", roomH.ClaimRoom)
			})
		})
	})

	return r
}
