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
	SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID, aiID uuid.UUID) (*domain.Message, error)
	GetLastMessage(ctx context.Context, roomID uuid.UUID) (*domain.Message, error)
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
	}, domain.EventNewQuestion, uuid.Nil, uuid.Nil)

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

func TestGetLastMessage_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "msg_human2", "Human")
	aiID := createTestUser(t, "msg_ai2", "AI")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)
	err = testRepo.ClaimRoom(context.Background(), room.ID, aiID)
	require.NoError(t, err)

	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: humanID, SenderRole: domain.Human, RoomID: room.ID,
		Content: "first", Tags: []string{"backend"},
	}, domain.EventNewQuestion, uuid.Nil, uuid.Nil)
	require.NoError(t, err)

	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: aiID, SenderRole: domain.AI, RoomID: room.ID,
		Content: "second", Tags: []string{},
	}, domain.EventAIReply, humanID, uuid.Nil)
	require.NoError(t, err)

	_, err = testRepo.SaveWithOutbox(context.Background(), &domain.Message{
		SenderID: humanID, SenderRole: domain.Human, RoomID: room.ID,
		Content: "third", Tags: []string{},
	}, domain.EventHumanFollowUp, uuid.Nil, aiID)
	require.NoError(t, err)

	msg, err := testRepo.GetLastMessage(context.Background(), room.ID)
	require.NoError(t, err)
	assert.Equal(t, "third", msg.Content)
	assert.Equal(t, domain.Human, msg.SenderRole)
}

func TestGetLastMessage_Empty(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "msg_human3", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	_, err = testRepo.GetLastMessage(context.Background(), room.ID)
	assert.ErrorIs(t, err, domain.ErrNoMessages)
}
