package errors

import (
	"net/http"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

var Registry = map[error]wrapper.ErrorResponse{
	domain.ErrMessageTooLong:              {Status: http.StatusBadRequest, Message: "Message content is too long"},
	domain.ErrMessageEmpty:                {Status: http.StatusBadRequest, Message: "Message content cannot be empty"},
	domain.ErrRateLimitExceeded:           {Status: http.StatusTooManyRequests, Message: "Please wait for an AI response before sending more messages"},
	domain.ErrInvalidTags:                 {Status: http.StatusBadRequest, Message: "Invalid tags"},
	domain.ErrAccessDenied:                {Status: http.StatusForbidden, Message: "Access denied"},
	domain.ErrAICannotStart:               {Status: http.StatusBadRequest, Message: "AI cannot start the conversation"},
	domain.ErrAIDoublePost:                {Status: http.StatusBadRequest, Message: "AI can only send one message at a time"},
	domain.ErrAuthorizationHeaderRequired: {Status: http.StatusUnauthorized, Message: "Authorization header is required"},
	domain.ErrInvalidAuthorizationFormat:  {Status: http.StatusUnauthorized, Message: "Invalid authorization format"},
	domain.ErrInvalidOrExpiredToken:       {Status: http.StatusUnauthorized, Message: "Invalid or expired token"},
	domain.ErrRoomNotFound:                {Status: http.StatusNotFound, Message: "Room not found"},
	domain.ErrRoomAlreadyClaimed:          {Status: http.StatusConflict, Message: "Room already claimed by another AI"},
	domain.ErrRoomAlreadyClosed:           {Status: http.StatusConflict, Message: "Room is already closed"},
	domain.ErrRoomNotActive:               {Status: http.StatusBadRequest, Message: "Room is not active"},
	domain.ErrNotRoomParticipant:          {Status: http.StatusForbidden, Message: "You are not a participant of this room"},
	domain.ErrRoomRequired:                {Status: http.StatusBadRequest, Message: "room_id is required for AI messages"},
}
