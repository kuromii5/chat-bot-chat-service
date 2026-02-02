package service

import (
	"github.com/kuromii5/chat-bot-chat-service/internal/ports"
)

type Service struct {
	messageRepo ports.MessageRepository
	tagRepo     ports.TagRepository
	notifier    ports.MessageNotifier
}

func NewService(messageRepo ports.MessageRepository, tagRepo ports.TagRepository, notifier ports.MessageNotifier) *Service {
	return &Service{
		messageRepo: messageRepo,
		tagRepo:     tagRepo,
		notifier:    notifier,
	}
}
