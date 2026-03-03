package room_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/service/room"
	"github.com/kuromii5/chat-bot-chat-service/internal/service/room/mocks"
)

var errDB = errors.New("db failure")

func TestClaimRoom(t *testing.T) {
	roomID := uuid.Must(uuid.NewV7())
	aiID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		setup   func(repo *mocks.MockRoomRepo)
		wantErr error
	}{
		{
			name: "success",
			setup: func(repo *mocks.MockRoomRepo) {
				repo.EXPECT().ClaimRoom(mock.Anything, roomID, aiID).Return(nil)
			},
		},
		{
			name: "repo error",
			setup: func(repo *mocks.MockRoomRepo) {
				repo.EXPECT().ClaimRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRoomRepo(t)
			svc := room.NewService(repo)

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.ClaimRoom(context.Background(), roomID, aiID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCloseRoom(t *testing.T) {
	roomID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		setup   func(repo *mocks.MockRoomRepo)
		wantErr error
	}{
		{
			name: "success",
			setup: func(repo *mocks.MockRoomRepo) {
				repo.EXPECT().CloseRoom(mock.Anything, roomID, userID).Return(nil)
			},
		},
		{
			name: "error",
			setup: func(repo *mocks.MockRoomRepo) {
				repo.EXPECT().CloseRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRoomRepo(t)
			svc := room.NewService(repo)

			if tt.setup != nil {
				tt.setup(repo)
			}

			err := svc.CloseRoom(context.Background(), roomID, userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
