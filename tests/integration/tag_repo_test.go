//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

type tagRepo interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) error
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

func TestUpdateProfileTags_Success(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai1", "AI")

	err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "frontend"})
	require.NoError(t, err)

	// Verify tags saved
	tags, err := testRepo.GetProfileTags(context.Background(), aiID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"backend", "frontend"}, tags)

	// Verify outbox event created in the same transaction
	var evt domain.OutboxEvent
	err = testDB.Get(&evt,
		`SELECT event_type, aggregate_id, payload FROM core.outbox_events
		 WHERE aggregate_id = $1 AND event_type = $2`,
		aiID, domain.EventTagsSync,
	)
	require.NoError(t, err)
	assert.Equal(t, domain.EventTagsSync, evt.EventType)
	assert.Equal(t, aiID, evt.AggregateID)

	var payload domain.TagSyncPayload
	require.NoError(t, json.Unmarshal(evt.Payload, &payload))
	assert.Equal(t, aiID, payload.UserID)
	assert.ElementsMatch(t, []string{"backend", "frontend"}, payload.Tags)
	assert.Empty(t, payload.OldTags)
}

func TestUpdateProfileTags_Overwrite(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai2", "AI")

	// Set initial tags
	err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "frontend"})
	require.NoError(t, err)

	// Overwrite with new tags
	err = testRepo.UpdateProfileTags(context.Background(), aiID, []string{"security", "databases"})
	require.NoError(t, err)

	// Verify only new tags remain
	tags, err := testRepo.GetProfileTags(context.Background(), aiID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"security", "databases"}, tags)

	// Verify second outbox event has correct old_tags
	var evts []domain.OutboxEvent
	err = testDB.Select(&evts,
		`SELECT event_type, payload FROM core.outbox_events
		 WHERE aggregate_id = $1 AND event_type = $2
		 ORDER BY created_at`,
		aiID, domain.EventTagsSync,
	)
	require.NoError(t, err)
	require.Len(t, evts, 2)

	var payload domain.TagSyncPayload
	require.NoError(t, json.Unmarshal(evts[1].Payload, &payload))
	assert.Equal(t, aiID, payload.UserID)
	assert.ElementsMatch(t, []string{"security", "databases"}, payload.Tags)
	assert.ElementsMatch(t, []string{"backend", "frontend"}, payload.OldTags)
}

func TestGetProfileTags_Success(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai3", "AI")

	err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "databases"})
	require.NoError(t, err)

	tags, err := testRepo.GetProfileTags(context.Background(), aiID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"backend", "databases"}, tags)
}

func TestGetProfileTags_Empty(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai4", "AI")

	tags, err := testRepo.GetProfileTags(context.Background(), aiID)
	require.NoError(t, err)
	assert.Empty(t, tags)
}
