package tag

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	apperrors "github.com/kuromii5/chat-bot-chat-service/internal/errors"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/middleware"
	"github.com/kuromii5/chat-bot-chat-service/internal/handlers/http/tag/mocks"
	"github.com/kuromii5/chat-bot-shared/validator"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

func TestMain(m *testing.M) {
	validator.Init()
	wrapper.RegisterErrors(apperrors.Registry)
	os.Exit(m.Run())
}

func withUser(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, userID)
}

// ─── UpdateProfileTags ───────────────────────────────────────────────────────

func TestUpdateProfileTags(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name       string
		body       string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name: "success",
			body: `{"tags":["go","backend"]}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().UpdateProfileTags(mock.Anything, userID, []string{"go", "backend"}).
					Return(nil)
			},
			wantStatus: http.StatusOK,
			wantKey:    "status",
			wantVal:    "success",
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid JSON body",
		},
		{
			name:       "validation: missing tags",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name:       "validation: empty tags array",
			body:       `{"tags":[]}`,
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Validation failed",
		},
		{
			name: "service: invalid tags",
			body: `{"tags":["nonexistent"]}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().UpdateProfileTags(mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("failed: %w", domain.ErrInvalidTags))
			},
			wantStatus: http.StatusBadRequest,
			wantKey:    "error",
			wantVal:    "Invalid tags",
		},
		{
			name: "service: generic error",
			body: `{"tags":["go"]}`,
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().UpdateProfileTags(mock.Anything, mock.Anything, mock.Anything).
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

			req := httptest.NewRequest(http.MethodPut, "/api/v1/chat/profile/tags", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(withUser(req.Context(), userID))
			rec := httptest.NewRecorder()

			h.UpdateProfileTags(rec, req)

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

// ─── GetProfileTags ──────────────────────────────────────────────────────────

func TestGetProfileTags(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name       string
		setup      func(svc *mocks.MockService)
		wantStatus int
		wantKey    string
		wantVal    any
	}{
		{
			name: "success",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetProfileTags(mock.Anything, userID).
					Return([]string{"go", "backend"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service: generic error",
			setup: func(svc *mocks.MockService) {
				svc.EXPECT().GetProfileTags(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("connection reset"))
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

			req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/profile/tags", nil)
			req = req.WithContext(withUser(req.Context(), userID))
			rec := httptest.NewRecorder()

			h.GetProfileTags(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if rec.Body.Len() > 0 {
				var body map[string]any
				err := json.NewDecoder(rec.Body).Decode(&body)
				assert.NoError(t, err)

				if tt.wantKey != "" {
					assert.Equal(t, tt.wantVal, body[tt.wantKey])
				}
			}
		})
	}
}
