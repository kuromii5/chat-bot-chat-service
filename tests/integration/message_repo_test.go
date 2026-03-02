//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type messageRepo interface {
	SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID uuid.UUID) (*domain.Message, error)
	GetLastMessages(ctx context.Context, roomID uuid.UUID, limit int) ([]*domain.Message, error)
}

func TestSaveWithOutbox_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "msg_human1", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	saved, err := testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID:   humanID,
		SenderRole: domain.Human,
		RoomID:     room.ID,
		Content:    "Hello, world!",
		Tags:       []string{"backend", "frontend"},
	}, domain.EventNewQuestion, uuid.Nil)

	require.NoError(t, err)
	assert.NotEmpty(t, saved.ID)
	assert.Equal(t, humanID, saved.SenderID)
	assert.Equal(t, domain.Human, saved.SenderRole)
	assert.Equal(t, room.ID, saved.RoomID)
	assert.Equal(t, "Hello, world!", saved.Content)
	assert.False(t, saved.CreatedAt.IsZero())

	// Verify outbox event was created
	var count int
	err = testDB.Get(&count, `SELECT COUNT(*) FROM core.outbox_events WHERE aggregate_id = $1`, saved.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetLastMessages_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "msg_human2", "Human")
	aiID := createTestUser(t, "msg_ai2", "AI")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)
	err = testRepo.ClaimRoom(context.Background(), room.ID, aiID)
	require.NoError(t, err)

	// Send 3 messages
	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: humanID, SenderRole: domain.Human, RoomID: room.ID,
		Content: "first", Tags: []string{"backend"},
	}, domain.EventNewQuestion, uuid.Nil)
	require.NoError(t, err)

	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: aiID, SenderRole: domain.AI, RoomID: room.ID,
		Content: "second", Tags: []string{},
	}, domain.EventAIReply, humanID)
	require.NoError(t, err)

	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: humanID, SenderRole: domain.Human, RoomID: room.ID,
		Content: "third", Tags: []string{},
	}, domain.EventFollowUp, uuid.Nil)
	require.NoError(t, err)

	// Get last 2 — should be in DESC order (third, second)
	msgs, err := testRepo.GetLastMessages(context.Background(), room.ID, 2)
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	assert.Equal(t, "third", msgs[0].Content)
	assert.Equal(t, "second", msgs[1].Content)
}

func TestGetLastMessages_Empty(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "msg_human3", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	msgs, err := testRepo.GetLastMessages(context.Background(), room.ID, 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)
}
