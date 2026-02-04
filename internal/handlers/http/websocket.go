package http

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kuromii5/chat-bot-chat-service/internal/ports"
	"github.com/sirupsen/logrus"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type NotificationHandler struct {
	notifier ports.MessageNotifier
}

func NewNotificationHandler(n ports.MessageNotifier) *NotificationHandler {
	return &NotificationHandler{notifier: n}
}

func (h *NotificationHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(UserIDKey).(uuid.UUID)
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
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
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
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	err = h.notifier.Listen(ctx, uid, func(ctx context.Context, body []byte) error {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteMessage(websocket.TextMessage, body)
	})

	if err != nil {
		logrus.WithError(err).Errorf("failed to start rabbitmq listener for user %s", uid)
		return
	}

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Info("websocket connection closed unexpectedly")
			}
			break
		}
	}
}
