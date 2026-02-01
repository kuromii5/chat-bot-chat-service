package http

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/service"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

type createMessageRequest struct {
	Content string   `json:"content" validate:"required,max=2048"`
	Tags    []string `json:"tags" validate:"required,max=5,min=1"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	userID, _ := r.Context().Value(UserIDKey).(uuid.UUID)
	userRole, _ := r.Context().Value(UserRoleKey).(domain.Role)

	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	msg, err := h.service.SendMessage(r.Context(), service.CreateMessageReq{
		UserID:  userID,
		Content: req.Content,
		Role:    userRole,
		Tags:    req.Tags,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusCreated, msg)
}
