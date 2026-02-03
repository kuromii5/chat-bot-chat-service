package domain

import "errors"

var (
	ErrMessageTooLong              = errors.New("message content is too long")
	ErrMessageEmpty                = errors.New("message content cannot be empty")
	ErrRateLimitExceeded           = errors.New("please wait for an AI response before sending more messages")
	ErrAICannotStart               = errors.New("AI cannot start the conversation")
	ErrAIDoublePost                = errors.New("AI can only send one message at a time")
	ErrAccessDenied                = errors.New("access denied")
	ErrInvalidTags                 = errors.New("invalid tags")
	ErrAuthorizationHeaderRequired = errors.New("authorization header is required")
	ErrInvalidAuthorizationFormat  = errors.New("invalid authorization format")
	ErrInvalidOrExpiredToken       = errors.New("invalid or expired token")
)
