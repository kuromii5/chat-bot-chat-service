package service

import (
	"github.com/kuromii5/chat-bot-chat-service/internal/ports"
)

type Service struct {
	messageRepo ports.MessageRepository
}

func NewService(messageRepo ports.MessageRepository) *Service {
	return &Service{
		messageRepo: messageRepo,
	}
}
