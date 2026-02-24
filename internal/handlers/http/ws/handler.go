package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Listener interface {
	Listen(ctx context.Context, userID uuid.UUID, handler func(ctx context.Context, body []byte) error) error
}

type Handler struct {
	listener Listener
}

func NewHandler(l Listener) *Handler {
	return &Handler{listener: l}
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		logrus.Error("user_id not found in context")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("failed to upgrade connection")
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			logrus.WithError(closeErr).Debug("websocket close")
		}
	}()

	if err := conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		logrus.WithError(err).Debug("set read deadline")
	}
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	err = h.listener.Listen(ctx, uid, func(ctx context.Context, body []byte) error {
		_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.TextMessage, body)
	})
	if err != nil {
		logrus.WithError(err).Errorf("failed to start rabbitmq listener for user %s", uid)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				logrus.WithError(err).Info("websocket connection closed unexpectedly")
			}
			break
		}
	}
}
