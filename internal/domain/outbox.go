package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventNewQuestion EventType = "new_question"
	EventFollowUp    EventType = "follow_up"
	EventAIReply     EventType = "ai_reply"
	EventTagsSync    EventType = "tags_sync"
	EventRoomClaimed EventType = "room_claimed"
)

type EventStatus string

const (
	EventPending    EventStatus = "pending"
	EventProcessing EventStatus = "processing"
	EventPublished  EventStatus = "published"
	EventDead       EventStatus = "dead"
)

type MessagePayload struct {
	Message *Message  `json:"message"`
	HumanID uuid.UUID `json:"human_id"`
}

type TagSyncPayload struct {
	UserID  uuid.UUID `json:"user_id"`
	Tags    []string  `json:"tags"`
	OldTags []string  `json:"old_tags"`
}

type RoomClaimedPayload struct {
	RoomID uuid.UUID `json:"room_id"`
	AiID   uuid.UUID `json:"ai_id"`
}

type OutboxEvent struct {
	ID            uuid.UUID       `db:"id"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   uuid.UUID       `db:"aggregate_id"`
	EventType     EventType       `db:"event_type"`
	Payload       json.RawMessage `db:"payload"`
	Status        EventStatus     `db:"status"`
	Attempts      *int            `db:"attempts"`
	MaxAttempts   *int            `db:"max_attempts"`
	NextRetryAt   time.Time       `db:"next_retry_at"`
	LastError     *string         `db:"last_error"`
	CreatedAt     time.Time       `db:"created_at"`
	PublishedAt   *time.Time      `db:"published_at"`
}
