package msg

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	msgservice "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

type createMessageRequest struct {
	Content string   `json:"content" validate:"required,max=2048"`
	Tags    []string `json:"tags"    validate:"required,max=5,min=1"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(domain.Role)

	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	saved, err := h.svc.SendMessage(r.Context(), msgservice.CreateMessageReq{
		UserID:  userID,
		Content: req.Content,
		Role:    userRole,
		Tags:    req.Tags,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusCreated, saved)
}
