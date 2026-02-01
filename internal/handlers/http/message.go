package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/service"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
	"github.com/kuromii5/chat-bot-chat-service/pkg/wrapper"
)

type createMessageRequest struct {
	Content string `json:"content" validate:"required,max=2048"`
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req createMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		wrapper.WrapError(w, r, errors.New("internal server error: unauthorized"))
		return
	}

	if err := validator.Validate(req); err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	msg, err := h.service.SendMessage(r.Context(), service.CreateMessageReq{
		UserID:  userID,
		Content: req.Content,
	})
	if err != nil {
		wrapper.WrapError(w, r, err)
		return
	}

	wrapper.JSON(w, http.StatusCreated, msg)
}
