//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

// waitForMessage blocks until handler receives a message or timeout fires.
func waitForMessage(t *testing.T, ch <-chan *domain.Message, timeout time.Duration) *domain.Message {
	t.Helper()
	select {
	case msg := <-ch:
		return msg
	case <-time.After(timeout):
		t.Fatal("timed out waiting for message")
		return nil
	}
}

// assertNoMessage asserts nothing arrives within the given duration.
func assertNoMessage(t *testing.T, ch <-chan *domain.Message, dur time.Duration) {
	t.Helper()
	select {
	case msg := <-ch:
		t.Fatalf("expected no message, got: %+v", msg)
	case <-time.After(dur):
		// ok
	}
}

func TestPublishNewQuestion_RoutedByTags(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())

	// AI subscribes to "backend" tag
	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"backend"}, nil)
	require.NoError(t, err)

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	// Publish question with matching tag
	question := &domain.Message{
		ID:         uuid.Must(uuid.NewV7()),
		SenderID:   uuid.Must(uuid.NewV7()),
		SenderRole: domain.Human,
		RoomID:     uuid.Must(uuid.NewV7()),
		Content:    "How do I use Go channels?",
		Tags:       []string{"backend"},
	}
	err = testRMQ.PublishNewQuestion(context.Background(), question)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, question.ID, got.ID)
	assert.Equal(t, question.Content, got.Content)
}

func TestPublishNewQuestion_NotRoutedWithoutMatchingTag(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())

	// AI subscribes to "frontend" only
	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"frontend"}, nil)
	require.NoError(t, err)

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	// Publish question with "security" tag — should NOT reach this AI
	err = testRMQ.PublishNewQuestion(context.Background(), &domain.Message{
		ID:      uuid.Must(uuid.NewV7()),
		RoomID:  uuid.Must(uuid.NewV7()),
		Content: "SQL injection help",
		Tags:    []string{"security"},
	})
	require.NoError(t, err)

	assertNoMessage(t, received, 500*time.Millisecond)
}

func TestPublishFollowUp_RoutedToClaimedAI(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())
	roomID := uuid.Must(uuid.NewV7())

	// Create AI queue and bind room to it
	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"backend"}, nil)
	require.NoError(t, err)
	err = testRMQ.BindRoomToAI(context.Background(), roomID, aiID)
	require.NoError(t, err)

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	followUp := &domain.Message{
		ID:      uuid.Must(uuid.NewV7()),
		RoomID:  roomID,
		Content: "Can you explain more?",
	}
	err = testRMQ.PublishFollowUp(context.Background(), roomID, followUp)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, followUp.ID, got.ID)
	assert.Equal(t, followUp.Content, got.Content)
}

func TestPublishAIReply_RoutedToHuman(t *testing.T) {
	humanID := uuid.Must(uuid.NewV7())

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Human starts listening for replies
	err := testRMQ.ListenReplies(ctx, humanID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	reply := &domain.Message{
		ID:      uuid.Must(uuid.NewV7()),
		RoomID:  uuid.Must(uuid.NewV7()),
		Content: "Here is the answer",
	}
	err = testRMQ.PublishAIReply(context.Background(), humanID, reply)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, reply.ID, got.ID)
	assert.Equal(t, reply.Content, got.Content)
}

func TestSyncAIQueue_UnbindsOldTags(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())

	// Initial: AI subscribes to "backend"
	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"backend"}, nil)
	require.NoError(t, err)

	// Update: switch from "backend" to "frontend"
	err = testRMQ.SyncAIQueue(context.Background(), aiID, []string{"frontend"}, []string{"backend"})
	require.NoError(t, err)

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	// Old tag "backend" — should NOT arrive
	err = testRMQ.PublishNewQuestion(context.Background(), &domain.Message{
		ID: uuid.Must(uuid.NewV7()), RoomID: uuid.Must(uuid.NewV7()),
		Content: "old tag question", Tags: []string{"backend"},
	})
	require.NoError(t, err)
	assertNoMessage(t, received, 500*time.Millisecond)

	// New tag "frontend" — SHOULD arrive
	frontendQ := &domain.Message{
		ID: uuid.Must(uuid.NewV7()), RoomID: uuid.Must(uuid.NewV7()),
		Content: "new tag question", Tags: []string{"frontend"},
	}
	err = testRMQ.PublishNewQuestion(context.Background(), frontendQ)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, frontendQ.ID, got.ID)
}

func TestPublishNewQuestion_MultiTagRouting(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())

	// AI subscribes to "databases" only
	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"databases"}, nil)
	require.NoError(t, err)

	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	// Publish question with multiple tags including "databases"
	question := &domain.Message{
		ID:      uuid.Must(uuid.NewV7()),
		RoomID:  uuid.Must(uuid.NewV7()),
		Content: "Postgres vs MySQL?",
		Tags:    []string{"backend", "databases"},
	}
	err = testRMQ.PublishNewQuestion(context.Background(), question)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, question.ID, got.ID)
}

func TestListen_NackRequeuesMessage(t *testing.T) {
	aiID := uuid.Must(uuid.NewV7())

	err := testRMQ.SyncAIQueue(context.Background(), aiID, []string{"backend"}, nil)
	require.NoError(t, err)

	attempt := 0
	received := make(chan *domain.Message, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = testRMQ.Listen(ctx, aiID, func(_ context.Context, body []byte) error {
		attempt++
		if attempt == 1 {
			// First attempt: fail — message should be requeued
			return assert.AnError
		}
		// Second attempt: succeed
		var msg domain.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			return err
		}
		received <- &msg
		return nil
	})
	require.NoError(t, err)

	question := &domain.Message{
		ID: uuid.Must(uuid.NewV7()), RoomID: uuid.Must(uuid.NewV7()),
		Content: "retry me", Tags: []string{"backend"},
	}
	err = testRMQ.PublishNewQuestion(context.Background(), question)
	require.NoError(t, err)

	got := waitForMessage(t, received, 5*time.Second)
	assert.Equal(t, question.ID, got.ID)
	assert.GreaterOrEqual(t, attempt, 2)
}
