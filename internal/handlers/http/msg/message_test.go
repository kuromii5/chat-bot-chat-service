package msg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/msg/mocks"
	msgservice "github.com/kuromii5/chat-bot-chat-service/internal/service/msg"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
)

func TestMain(m *testing.M) {
	validator.Init()
	os.Exit(m.Run())
}

func withUser(ctx context.Context, userID uuid.UUID, role domain.Role) context.Context {
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, role)
	return ctx
}

func TestSendMessage(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name       string
		body       string
		userID     uuid.UUID
		role       domain.Role
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name:   "success",
			body:   fmt.Sprintf(`{"content":"hello","tags":["go"],"room_id":"%s"}`, roomID),
			userID: userID,
			role:   domain.Human,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(&domain.Message{
						ID: uuid.Must(uuid.NewV7()), SenderID: userID,
						SenderRole: domain.Human, RoomID: roomID,
						Content: "hello", CreatedAt: time.Now(),
					}, nil)
			},
			wantStatus: http.StatusCreated,
			wantKey:    "Content",
			wantVal:    "hello",
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			userID:     userID,
			role:       domain.Human,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing content",
			body:       `{"tags":["go"]}`,
			userID:     userID,
			role:       domain.Human,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: content too long",
			body:       `{"content":"` + strings.Repeat("x", 2049) + `"}`,
			userID:     userID,
			role:       domain.Human,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: too many tags",
			body:       `{"content":"hello","tags":["a","b","c","d","e","f"]}`,
			userID:     userID,
			role:       domain.Human,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:   "service: room required",
			body:   `{"content":"hello"}`,
			userID: userID,
			role:   domain.AI,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrRoomRequired))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "room_id is required for AI messages",
		},
		{
			name:   "service: room not active",
			body:   fmt.Sprintf(`{"content":"hello","room_id":"%s"}`, roomID),
			userID: userID,
			role:   domain.AI,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrRoomNotActive))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Room is not active",
		},
		{
			name:   "service: not participant",
			body:   fmt.Sprintf(`{"content":"hello","room_id":"%s"}`, roomID),
			userID: userID,
			role:   domain.AI,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrNotRoomParticipant))
			},
			wantStatus: http.StatusForbidden,
			wantKey:    "error",
			wantVal:    "You are not a participant of this room",
		},
		{
			name:   "service: AI double post",
			body:   fmt.Sprintf(`{"content":"hello","room_id":"%s"}`, roomID),
			userID: userID,
			role:   domain.AI,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrAIDoublePost))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "AI can only send one message at a time",
		},
		{
			name:   "service: AI cannot start",
			body:   fmt.Sprintf(`{"content":"hello","room_id":"%s"}`, roomID),
			userID: userID,
			role:   domain.AI,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrAICannotStart))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "AI cannot start the conversation",
		},
		{
			name:   "service: invalid tags",
			body:   `{"content":"hello","tags":["nonexistent"]}`,
			userID: userID,
			role:   domain.Human,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("failed: %w", domain.ErrInvalidTags))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid tags",
		},
		{
			name:   "service: generic error",
			body:   `{"content":"hello","tags":["go"]}`,
			userID: userID,
			role:   domain.Human,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("connection reset"))
			},
			wantStatus: http.StatusInternalServerError,
			wantKey:    "error",
			wantVal:    "Internal server error",
		},
		{
			name:   "passes correct userID and role to service",
			body:   `{"content":"hello","tags":["go"]}`,
			userID: userID,
			role:   domain.Human,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().SendMessage(mock.Anything, mock.MatchedBy(func(req msgservice.CreateMessageReq) bool {
					return req.UserID == userID && req.Role == domain.Human
				})).Return(&domain.Message{}, nil)
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/send", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(withUser(req.Context(), tt.userID, tt.role))
			rec := httptest.NewRecorder()

			h.SendMessage(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantKey != "" && rec.Body.Len() > 0 {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVal, body[tt.wantKey])
			}
		})
	}
}
