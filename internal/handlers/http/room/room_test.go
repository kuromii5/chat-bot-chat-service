package room

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/room/mocks"
)

func withUserAndRoom(ctx context.Context, userID uuid.UUID, roomID string) context.Context {
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("roomID", roomID)
	return context.WithValue(ctx, chi.RouteCtxKey, rctx)
}

// ─── ClaimRoom ───────────────────────────────────────────────────────────────

func TestClaimRoom(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name       string
		roomParam  string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name:      "success",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ClaimRoom(mock.Anything, roomID, aiID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid room id",
			roomParam:  "not-a-uuid",
			wantStatus: http.StatusNotFound,
			wantKey:    "error",
			wantVal:    "Room not found",
		},
		{
			name:      "service: room not found",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ClaimRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("failed: %w", domain.ErrRoomNotFound))
			},
			wantStatus: http.StatusNotFound,
			wantKey:    "error",
			wantVal:    "Room not found",
		},
		{
			name:      "service: room already claimed",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ClaimRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("failed: %w", domain.ErrRoomAlreadyClaimed))
			},
			wantStatus: http.StatusConflict,
			wantKey:    "error",
			wantVal:    "Room already claimed by another AI",
		},
		{
			name:      "service: generic error",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().ClaimRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("connection reset"))
			},
			wantStatus: http.StatusInternalServerError,
			wantKey:    "error",
			wantVal:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/rooms/"+tt.roomParam+"/claim", nil)
			req = req.WithContext(withUserAndRoom(req.Context(), aiID, tt.roomParam))
			rec := httptest.NewRecorder()

			h.ClaimRoom(rec, req)

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

// ─── CloseRoom ───────────────────────────────────────────────────────────────

func TestCloseRoom(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name       string
		roomParam  string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name:      "success",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().CloseRoom(mock.Anything, roomID, userID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid room id",
			roomParam:  "not-a-uuid",
			wantStatus: http.StatusNotFound,
			wantKey:    "error",
			wantVal:    "Room not found",
		},
		{
			name:      "service: room already closed",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().CloseRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("failed: %w", domain.ErrRoomAlreadyClosed))
			},
			wantStatus: http.StatusConflict,
			wantKey:    "error",
			wantVal:    "Room is already closed",
		},
		{
			name:      "service: not participant",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().CloseRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("failed: %w", domain.ErrNotRoomParticipant))
			},
			wantStatus: http.StatusForbidden,
			wantKey:    "error",
			wantVal:    "You are not a participant of this room",
		},
		{
			name:      "service: generic error",
			roomParam: roomID.String(),
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().CloseRoom(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("connection reset"))
			},
			wantStatus: http.StatusInternalServerError,
			wantKey:    "error",
			wantVal:    "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockService(t)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewHandler(svc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/rooms/"+tt.roomParam+"/close", nil)
			req = req.WithContext(withUserAndRoom(req.Context(), userID, tt.roomParam))
			rec := httptest.NewRecorder()

			h.CloseRoom(rec, req)

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
