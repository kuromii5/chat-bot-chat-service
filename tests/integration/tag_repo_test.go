//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type tagRepo interface {
	UpdateProfileTags(ctx context.Context, userID uuid.UUID, tags []string) (oldTags []string, err error)
	GetProfileTags(ctx context.Context, userID uuid.UUID) ([]string, error)
}

func TestUpdateProfileTags_Success(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai1", "AI")

	oldTags, err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "frontend"})

	require.NoError(t, err)
	assert.Empty(t, oldTags) // first time — no old tags
}

func TestUpdateProfileTags_ReturnsOldTags(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai2", "AI")

	// Set initial tags (must exist in core.tags from migration 008)
	_, err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "frontend"})
	require.NoError(t, err)

	// Update to new tags — should return old ones
	oldTags, err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"security", "databases"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"backend", "frontend"}, oldTags)
}

func TestGetProfileTags_Success(t *testing.T) {
	truncateAll(t)
	aiID := createTestUser(t, "tag_ai3", "AI")

	_, err := testRepo.UpdateProfileTags(context.Background(), aiID, []string{"backend", "databases"})
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
