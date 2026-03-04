package msg_test

import (
	"context"
	"errors"
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
	"github.com/kuromii5/chat-bot-shared/validator"
)

var errDB = errors.New("db failure")

func TestMain(m *testing.M) {
	validator.Init()
	os.Exit(m.Run())
}

func TestSendMessage_Human_NewQuestion(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		req     msg.CreateMessageReq
		setup   func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo)
		wantErr error
		check   func(t *testing.T, result *domain.Message)
	}{
		{
			name: "success",
			req: msg.CreateMessageReq{
				UserID:  userID,
				Role:    domain.Human,
				Content: "What is the best way to learn Go?",
				Tags:    []string{"backend"},
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().
					CreateRoom(mock.Anything, userID).
					Return(&domain.Room{ID: roomID, HumanID: userID, Status: domain.RoomOpen}, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
						return m.SenderID == userID &&
							m.SenderRole == domain.Human &&
							m.RoomID == roomID
					}), domain.EventNewQuestion, uuid.Nil, uuid.Nil).
					Return(&domain.Message{
						ID: uuid.Must(uuid.NewV7()), SenderID: userID, SenderRole: domain.Human,
						RoomID: roomID, Content: "What is the best way to learn Go?",
						Tags: []string{"backend"}, CreatedAt: time.Now(),
					}, nil)
			},
			check: func(t *testing.T, result *domain.Message) {
				assert.Equal(t, userID, result.SenderID)
				assert.Equal(t, roomID, result.RoomID)
			},
		},
		{
			name: "empty tags",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "question", Tags: []string{},
			},
			wantErr: domain.ErrInvalidTags,
		},
		{
			name: "create room error",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "question", Tags: []string{"go"},
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().
					CreateRoom(mock.Anything, mock.Anything).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "save error",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "question", Tags: []string{"go"},
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().
					CreateRoom(mock.Anything, mock.Anything).
					Return(&domain.Room{ID: roomID}, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.Anything, domain.EventNewQuestion, uuid.Nil, uuid.Nil).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "deduplicates and sorts tags",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "question",
				Tags: []string{"go", "backend", "go", "backend", "api"},
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().
					CreateRoom(mock.Anything, mock.Anything).
					Return(&domain.Room{ID: roomID}, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
						return len(m.Tags) == 3 &&
							m.Tags[0] == "api" &&
							m.Tags[1] == "backend" &&
							m.Tags[2] == "go"
					}), domain.EventNewQuestion, uuid.Nil, uuid.Nil).
					Return(&domain.Message{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMessageRepo(t)
			roomRepo := mocks.NewMockRoomRepo(t)
			svc := msg.NewService(repo, roomRepo)

			if tt.setup != nil {
				tt.setup(repo, roomRepo)
			}

			result, err := svc.SendMessage(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestSendMessage_Human_FollowUp(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	aiID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())
	otherUserID := uuid.Must(uuid.NewV7())

	activeRoom := &domain.Room{ID: roomID, HumanID: userID, AIID: &aiID, Status: domain.RoomActive}

	tests := []struct {
		name    string
		req     msg.CreateMessageReq
		setup   func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo)
		wantErr error
		check   func(t *testing.T, result *domain.Message)
	}{
		{
			name: "success",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "follow up", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
						return m.SenderID == userID && m.RoomID == roomID && m.SenderRole == domain.Human
					}), domain.EventHumanFollowUp, uuid.Nil, aiID).
					Return(&domain.Message{
						ID: uuid.Must(uuid.NewV7()), SenderID: userID,
						SenderRole: domain.Human, RoomID: roomID, Content: "follow up",
					}, nil)
			},
			check: func(t *testing.T, result *domain.Message) {
				assert.Equal(t, userID, result.SenderID)
				assert.Equal(t, roomID, result.RoomID)
			},
		},
		{
			name: "get room error",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "follow up", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, mock.Anything).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "room not active",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "follow up", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).
					Return(&domain.Room{ID: roomID, Status: domain.RoomClosed}, nil)
			},
			wantErr: domain.ErrRoomNotActive,
		},
		{
			name: "not participant",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "follow up", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).
					Return(&domain.Room{ID: roomID, HumanID: otherUserID, Status: domain.RoomActive}, nil)
			},
			wantErr: domain.ErrNotRoomParticipant,
		},
		{
			name: "save error",
			req: msg.CreateMessageReq{
				UserID: userID, Role: domain.Human, Content: "follow up", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.Anything, domain.EventHumanFollowUp, uuid.Nil, aiID).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMessageRepo(t)
			roomRepo := mocks.NewMockRoomRepo(t)
			svc := msg.NewService(repo, roomRepo)

			if tt.setup != nil {
				tt.setup(repo, roomRepo)
			}

			result, err := svc.SendMessage(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestSendMessage_AI_Reply(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())
	humanID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())
	otherAI := uuid.Must(uuid.NewV7())

	activeRoom := &domain.Room{
		ID: roomID, HumanID: humanID, AIID: &aiID, Status: domain.RoomActive,
	}
	humanLastMsg := &domain.Message{SenderRole: domain.Human}
	aiLastMsg := &domain.Message{SenderRole: domain.AI}

	tests := []struct {
		name    string
		req     msg.CreateMessageReq
		setup   func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo)
		wantErr error
		check   func(t *testing.T, result *domain.Message)
	}{
		{
			name: "success",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "Here is the answer", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					GetLastMessage(mock.Anything, roomID).
					Return(humanLastMsg, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.MatchedBy(func(m *domain.Message) bool {
						return m.SenderID == aiID && m.SenderRole == domain.AI && m.RoomID == roomID
					}), domain.EventAIReply, humanID, uuid.Nil).
					Return(&domain.Message{
						ID: uuid.Must(uuid.NewV7()), SenderID: aiID,
						SenderRole: domain.AI, RoomID: roomID, Content: "Here is the answer",
					}, nil)
			},
			check: func(t *testing.T, result *domain.Message) {
				assert.Equal(t, aiID, result.SenderID)
				assert.Equal(t, domain.AI, result.SenderRole)
			},
		},
		{
			name: "no room id",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer",
			},
			wantErr: domain.ErrRoomRequired,
		},
		{
			name: "get room error",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, mock.Anything).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "room not active",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).
					Return(&domain.Room{ID: roomID, AIID: &aiID, Status: domain.RoomOpen}, nil)
			},
			wantErr: domain.ErrRoomNotActive,
		},
		{
			name: "not participant - nil AI ID",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).
					Return(&domain.Room{ID: roomID, AIID: nil, Status: domain.RoomActive}, nil)
			},
			wantErr: domain.ErrNotRoomParticipant,
		},
		{
			name: "not participant - different AI",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(_ *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).
					Return(&domain.Room{ID: roomID, AIID: &otherAI, Status: domain.RoomActive}, nil)
			},
			wantErr: domain.ErrNotRoomParticipant,
		},
		{
			name: "cannot start conversation",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					GetLastMessage(mock.Anything, roomID).
					Return(nil, domain.ErrNoMessages)
			},
			wantErr: domain.ErrAICannotStart,
		},
		{
			name: "double post",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					GetLastMessage(mock.Anything, roomID).
					Return(aiLastMsg, nil)
			},
			wantErr: domain.ErrAIDoublePost,
		},
		{
			name: "get last message error",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					GetLastMessage(mock.Anything, roomID).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "save error",
			req: msg.CreateMessageReq{
				UserID: aiID, Role: domain.AI, Content: "answer", RoomID: roomID,
			},
			setup: func(repo *mocks.MockMessageRepo, roomRepo *mocks.MockRoomRepo) {
				roomRepo.EXPECT().GetRoom(mock.Anything, roomID).Return(activeRoom, nil)
				repo.EXPECT().
					GetLastMessage(mock.Anything, roomID).
					Return(humanLastMsg, nil)
				repo.EXPECT().
					SaveWithOutbox(mock.Anything, mock.Anything, domain.EventAIReply, mock.Anything, uuid.Nil).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockMessageRepo(t)
			roomRepo := mocks.NewMockRoomRepo(t)
			svc := msg.NewService(repo, roomRepo)

			if tt.setup != nil {
				tt.setup(repo, roomRepo)
			}

			result, err := svc.SendMessage(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestSendMessage_UnknownRole(t *testing.T) {
	repo := mocks.NewMockMessageRepo(t)
	roomRepo := mocks.NewMockRoomRepo(t)
	svc := msg.NewService(repo, roomRepo)

	result, err := svc.SendMessage(context.Background(), msg.CreateMessageReq{
		UserID: uuid.Must(uuid.NewV7()), Role: "unknown", Content: "test",
	})

	assert.ErrorIs(t, err, domain.ErrAccessDenied)
	assert.Nil(t, result)
}
