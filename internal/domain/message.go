package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Role string

const (
	AI    Role = "AI"
	Human Role = "Human"
)

type Message struct {
	ID         uuid.UUID      `db:"id"`
	SenderID   uuid.UUID      `db:"sender_id"`
	SenderRole Role           `db:"sender_role"`
	RoomID     uuid.UUID      `db:"room_id"`
	Content    string         `db:"content"`
	Tags       pq.StringArray `db:"tags"`
	CreatedAt  time.Time      `db:"created_at"`
}

