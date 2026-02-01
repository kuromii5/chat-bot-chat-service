package http

import "github.com/kuromii5/chat-bot-chat-service/internal/service"

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}
