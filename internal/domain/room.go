package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomOpen   RoomStatus = "open"
	RoomActive RoomStatus = "active"
	RoomClosed RoomStatus = "closed"
)

type Room struct {
	ID        uuid.UUID  `db:"id"`
	HumanID   uuid.UUID  `db:"human_id"`
	AIID      *uuid.UUID `db:"ai_id"`
	Status    RoomStatus `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	ClosedAt  *time.Time `db:"closed_at"`
}
