package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	AI    Role = "AI"
	Human Role = "Human"
)

const (
	HumanSequentialMessageLimit = 3
	AISequentialMessageLimit    = 1
)

type Message struct {
	ID         uuid.UUID  `db:"id"`
	SenderID   uuid.UUID  `db:"sender_id"`
	SenderRole Role       `db:"sender_role"`
	RoomID     string     `db:"room_id"`
	Content    string     `db:"content"`
	Tags       []string   `db:"tags"`
	CreatedAt  time.Time  `db:"created_at"`
	AssignedTo *uuid.UUID `db:"assigned_to"`
}

func ValidateHumanMsg(lastMessages []*Message) error {
	if len(lastMessages) < HumanSequentialMessageLimit {
		return nil
	}

	humanCount := 0
	for i := range HumanSequentialMessageLimit {
		if lastMessages[i].SenderRole == Human {
			humanCount++
		} else {
			return nil
		}
	}

	if humanCount >= HumanSequentialMessageLimit {
		return ErrRateLimitExceeded
	}
	return nil
}

func ValidateAIMsg(lastMessages []*Message) error {
	if len(lastMessages) < AISequentialMessageLimit {
		return ErrAICannotStart
	}

	lastMsg := lastMessages[0]
	if lastMsg.SenderRole == AI {
		return ErrAIDoublePost
	}

	return nil
}
