package msg_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
	"github.com/kuromii5/chat-bot-chat-service/internal/service/msg/mocks"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

func TestMain(m *testing.M) {
	validator.Init()
	os.Exit(m.Run())
}

func TestSendMessage_Human_NewQuestion_Success(t *testing.T) {
	repo := mocks.NewMockMessageRepo(t)
	roomRepo := mocks.NewMockRoomRepo(t)
	svc := msg.NewService(repo, roomRepo)

	userID, _ := uuid.NewV7()
	roomID, _ := uuid.NewV7()
	req := msg.CreateMessageReq{
		UserID:  userID,
		Content: "What is the best way to learn Go?",
		Role:    domain.Human,
		Tags:    []string{"backend"},
	}

	createdRoom := &domain.Room{
		ID:      roomID,
		HumanID: userID,
		Status:  domain.RoomOpen,
	}
	savedMsg := &domain.Message{
		ID:         uuid.Must(uuid.NewV7()),
		SenderID:   userID,
		SenderRole: domain.Human,
		RoomID:     roomID,
		Content:    req.Content,
		Tags:       req.Tags,
		CreatedAt:  time.Now(),
	}

	roomRepo.EXPECT().
		CreateRoom(mock.Anything, userID).
		Return(createdRoom, nil)
	repo.EXPECT().
		SaveWithOutbox(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
			return m.SenderID == userID &&
				m.SenderRole == domain.Human &&
				m.RoomID == roomID &&
				m.Content == req.Content
		}), domain.EventNewQuestion, uuid.Nil).
		Return(savedMsg, nil)

	result, err := svc.SendMessage(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, savedMsg, result)
}
