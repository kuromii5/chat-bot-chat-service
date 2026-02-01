package wrapper

import (
	"net/http"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type ErrorResponse struct {
	Status  int
	Message string
}

var errorRegistry = map[error]ErrorResponse{
	domain.ErrMessageTooLong: {
		Status:  http.StatusBadRequest,
		Message: "Message content is too long",
	},
	domain.ErrMessageEmpty: {
		Status:  http.StatusBadRequest,
		Message: "Message content cannot be empty",
	},
	domain.ErrRateLimitExceeded: {
		Status:  http.StatusTooManyRequests,
		Message: "Please wait for an AI response before sending more messages",
	},
}
