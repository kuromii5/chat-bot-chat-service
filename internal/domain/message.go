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
