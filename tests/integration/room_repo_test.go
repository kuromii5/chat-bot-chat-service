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

type roomRepo interface {
	CreateRoom(ctx context.Context, humanID uuid.UUID) (*domain.Room, error)
	ClaimRoom(ctx context.Context, roomID uuid.UUID, aiID uuid.UUID) error
	CloseRoom(ctx context.Context, roomID uuid.UUID, userID uuid.UUID) error
	GetRoom(ctx context.Context, roomID uuid.UUID) (*domain.Room, error)
}

func TestCreateRoom_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human1", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)

	require.NoError(t, err)
	assert.NotEmpty(t, room.ID)
	assert.Equal(t, humanID, room.HumanID)
	assert.Nil(t, room.AIID)
	assert.Equal(t, domain.RoomOpen, room.Status)
	assert.False(t, room.CreatedAt.IsZero())
}

func TestClaimRoom_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human2", "Human")
	aiID := createTestUser(t, "ai2", "AI")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	err = testRepo.ClaimRoom(context.Background(), room.ID, aiID)
	require.NoError(t, err)

	// Verify room is now active with AI assigned
	claimed, err := testRepo.GetRoom(context.Background(), room.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RoomActive, claimed.Status)
	assert.NotNil(t, claimed.AIID)
	assert.Equal(t, aiID, *claimed.AIID)
}

func TestClaimRoom_AlreadyClaimed(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human3", "Human")
	ai1 := createTestUser(t, "ai3a", "AI")
	ai2 := createTestUser(t, "ai3b", "AI")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	err = testRepo.ClaimRoom(context.Background(), room.ID, ai1)
	require.NoError(t, err)

	// Second claim should fail
	err = testRepo.ClaimRoom(context.Background(), room.ID, ai2)
	assert.ErrorIs(t, err, domain.ErrRoomAlreadyClaimed)
}

func TestCloseRoom_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human4", "Human")
	aiID := createTestUser(t, "ai4", "AI")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)
	err = testRepo.ClaimRoom(context.Background(), room.ID, aiID)
	require.NoError(t, err)

	err = testRepo.CloseRoom(context.Background(), room.ID, humanID)
	require.NoError(t, err)

	closed, err := testRepo.GetRoom(context.Background(), room.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RoomClosed, closed.Status)
	assert.NotNil(t, closed.ClosedAt)
}

func TestCloseRoom_NotFound(t *testing.T) {
	truncateAll(t)

	err := testRepo.CloseRoom(context.Background(), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))
	assert.ErrorIs(t, err, domain.ErrRoomNotFound)
}

func TestCloseRoom_NotParticipant(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human5", "Human")
	stranger := createTestUser(t, "stranger5", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	err = testRepo.CloseRoom(context.Background(), room.ID, stranger)
	assert.ErrorIs(t, err, domain.ErrNotRoomParticipant)
}

func TestCloseRoom_AlreadyClosed(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human6", "Human")

	room, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	err = testRepo.CloseRoom(context.Background(), room.ID, humanID)
	require.NoError(t, err)

	// Second close should fail
	err = testRepo.CloseRoom(context.Background(), room.ID, humanID)
	assert.ErrorIs(t, err, domain.ErrRoomAlreadyClosed)
}

func TestGetRoom_Success(t *testing.T) {
	truncateAll(t)
	humanID := createTestUser(t, "human7", "Human")

	created, err := testRepo.CreateRoom(context.Background(), humanID)
	require.NoError(t, err)

	found, err := testRepo.GetRoom(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, humanID, found.HumanID)
}

func TestGetRoom_NotFound(t *testing.T) {
	truncateAll(t)

	_, err := testRepo.GetRoom(context.Background(), uuid.Must(uuid.NewV7()))
	assert.ErrorIs(t, err, domain.ErrRoomNotFound)
}
