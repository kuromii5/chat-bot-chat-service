package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/kuromii5/chat-bot-chat-service/internal/service/tag"
	"github.com/kuromii5/chat-bot-chat-service/internal/service/tag/mocks"
)

var errDB = errors.New("db failure")

func TestUpdateProfileTags(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name    string
		tags    []string
		setup   func(repo *mocks.MockTagRepo, cache *mocks.MockTagCache, notifier *mocks.MockNotifier)
		wantErr error
	}{
		{
			name: "success",
			tags: []string{"go", "backend"},
			setup: func(repo *mocks.MockTagRepo, cache *mocks.MockTagCache, notifier *mocks.MockNotifier) {
				cache.EXPECT().AreTagsValid(mock.Anything, []string{"backend", "go"}).Return(true)
				repo.EXPECT().
					UpdateProfileTags(mock.Anything, userID, []string{"backend", "go"}).
					Return([]string{"python"}, nil)
				notifier.EXPECT().
					SyncAIQueue(mock.Anything, userID, []string{"backend", "go"}, []string{"python"}).
					Return(nil)
			},
		},
		{
			name: "invalid tags",
			tags: []string{"nonexistent"},
			setup: func(_ *mocks.MockTagRepo, cache *mocks.MockTagCache, _ *mocks.MockNotifier) {
				cache.EXPECT().AreTagsValid(mock.Anything, mock.Anything).Return(false)
			},
			wantErr: domain.ErrInvalidTags,
		},
		{
			name: "repo error",
			tags: []string{"go"},
			setup: func(repo *mocks.MockTagRepo, cache *mocks.MockTagCache, _ *mocks.MockNotifier) {
				cache.EXPECT().AreTagsValid(mock.Anything, mock.Anything).Return(true)
				repo.EXPECT().
					UpdateProfileTags(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "notifier sync error",
			tags: []string{"go"},
			setup: func(repo *mocks.MockTagRepo, cache *mocks.MockTagCache, notifier *mocks.MockNotifier) {
				cache.EXPECT().AreTagsValid(mock.Anything, mock.Anything).Return(true)
				repo.EXPECT().
					UpdateProfileTags(mock.Anything, mock.Anything, mock.Anything).
					Return([]string{}, nil)
				notifier.EXPECT().
					SyncAIQueue(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errDB)
			},
			wantErr: errDB,
		},
		{
			name: "sorts and deduplicates tags",
			tags: []string{"go", "backend", "go", "api", "backend"},
			setup: func(repo *mocks.MockTagRepo, cache *mocks.MockTagCache, notifier *mocks.MockNotifier) {
				cache.EXPECT().
					AreTagsValid(mock.Anything, mock.MatchedBy(func(tags []string) bool {
						return len(tags) == 3 &&
							tags[0] == "api" && tags[1] == "backend" && tags[2] == "go"
					})).
					Return(true)
				repo.EXPECT().
					UpdateProfileTags(mock.Anything, mock.Anything, mock.MatchedBy(func(tags []string) bool {
						return len(tags) == 3 &&
							tags[0] == "api" && tags[1] == "backend" && tags[2] == "go"
					})).
					Return([]string{}, nil)
				notifier.EXPECT().
					SyncAIQueue(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTagRepo(t)
			cache := mocks.NewMockTagCache(t)
			notifier := mocks.NewMockNotifier(t)
			svc := tag.NewService(repo, cache, notifier)

			if tt.setup != nil {
				tt.setup(repo, cache, notifier)
			}

			err := svc.UpdateProfileTags(context.Background(), userID, tt.tags)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGetProfileTags(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())

	tests := []struct {
		name     string
		setup    func(repo *mocks.MockTagRepo)
		wantTags []string
		wantErr  error
	}{
		{
			name: "success",
			setup: func(repo *mocks.MockTagRepo) {
				repo.EXPECT().GetProfileTags(mock.Anything, userID).
					Return([]string{"go", "backend"}, nil)
			},
			wantTags: []string{"go", "backend"},
		},
		{
			name: "error",
			setup: func(repo *mocks.MockTagRepo) {
				repo.EXPECT().GetProfileTags(mock.Anything, mock.Anything).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockTagRepo(t)
			svc := tag.NewService(repo, nil, nil)

			if tt.setup != nil {
				tt.setup(repo)
			}

			result, err := svc.GetProfileTags(context.Background(), userID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTags, result)
		})
	}
}
