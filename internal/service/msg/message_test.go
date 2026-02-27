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

func TestSendMessage_Human_Success(t *testing.T) {
	repo := mocks.NewMockMessageRepo(t)
	notifier := mocks.NewMockNotifier(t)
	svc := msg.NewService(repo, notifier)

	userID, _ := uuid.NewV7()
	req := msg.CreateMessageReq{
		UserID:  userID,
		Content: "What is the best way to learn Go?",
		Role:    domain.Human,
		Tags:    []string{"backend"},
	}

	savedMsg := &domain.Message{
		ID:         uuid.New(),
		SenderID:   userID,
		SenderRole: domain.Human,
		RoomID:     "global",
		Content:    req.Content,
		Tags:       req.Tags,
		CreatedAt:  time.Now(),
	}

	repo.EXPECT().
		GetLastMessages(mock.Anything, "global", domain.HumanSequentialMessageLimit).
		Return([]*domain.Message{}, nil)
	repo.EXPECT().
		Save(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
			return m.SenderID == userID &&
				m.SenderRole == domain.Human &&
				m.Content == req.Content &&
				m.RoomID == "global"
		})).
		Return(savedMsg, nil)

	published := make(chan struct{})
	notifier.EXPECT().
		PublishNewQuestion(mock.Anything, savedMsg).
		Run(func(_ context.Context, _ *domain.Message) {
			close(published)
		}).
		Return(nil)

	result, err := svc.SendMessage(context.Background(), req)

	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("PublishNewQuestion was not called within 1s")
	}

	require.NoError(t, err)
	assert.Equal(t, savedMsg, result)
}
