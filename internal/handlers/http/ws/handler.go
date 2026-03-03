package ws

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-shared/wrapper"
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
	ListenReplies(ctx context.Context, userID uuid.UUID, handler func(ctx context.Context, body []byte) error) error
}

type Handler struct {
	listener Listener
}

func NewHandler(l Listener) *Handler {
	return &Handler{listener: l}
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	uid, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	role, _ := r.Context().Value(middleware.UserRoleKey).(domain.Role)
	expiry, ok := r.Context().Value(middleware.TokenExpiryKey).(time.Time)
	if !ok || expiry.IsZero() {
		wrapper.WrapError(w, r, domain.ErrInvalidOrExpiredToken)
		return
	}

	var listenFn func(context.Context, uuid.UUID, func(context.Context, []byte) error) error
	switch role {
	case domain.AI:
		listenFn = h.listener.Listen
	case domain.Human:
		listenFn = h.listener.ListenReplies
	default:
		w.WriteHeader(http.StatusForbidden)
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

	ctx, cancel := context.WithDeadline(r.Context(), expiry)
	defer cancel()

	go func() {
		<-ctx.Done()
		_ = conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "token expired"),
			time.Now().Add(writeWait),
		)
		conn.Close()
	}()

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

	err = listenFn(ctx, uid, func(ctx context.Context, body []byte) error {
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
