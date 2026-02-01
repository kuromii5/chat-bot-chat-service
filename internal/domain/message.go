package domain

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID         uuid.UUID
	SenderID   uuid.UUID
	SenderRole string
	RoomID     string
	Content    string
	CreatedAt  time.Time
}

func CanUserSendMessage(lastMessages []*Message) bool {
	if len(lastMessages) < 3 {
		return true
	}

	humanCounter := 0
	for _, msg := range lastMessages {
		if msg.SenderRole == "AI" {
			return true
		}
		if msg.SenderRole == "Human" {
			humanCounter++
		}

		if humanCounter >= 3 {
			return false
		}
	}

	return true
}
