package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/kuromii5/chat-bot-chat-service/internal/adapters/outbox/mocks"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func newRelay(t *testing.T) (
	*Relay,
	*mocks.MockOutboxRepo,
	*mocks.MockPublisher,
	*mocks.MockQueueSyncer,
	*mocks.MockBinder,
) {
	t.Helper()
	repo := mocks.NewMockOutboxRepo(t)
	pub := mocks.NewMockPublisher(t)
	syncer := mocks.NewMockQueueSyncer(t)
	binder := mocks.NewMockBinder(t)
	relay := NewRelay(repo, pub, syncer, binder, time.Hour)
	return relay, repo, pub, syncer, binder
}

func TestProcess_FetchError(t *testing.T) {
	relay, repo, _, _, _ := newRelay(t)
	ctx := context.Background()
	// fetch fails — no panic, error is swallowed internally
	repo.EXPECT().FetchPending(ctx, fetchLimit).Return(nil, errors.New("db error"))
	relay.process(ctx)
}

func TestProcess_NewQuestion(t *testing.T) {
	relay, repo, pub, _, _ := newRelay(t)
	ctx := context.Background()

	msg := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventNewQuestion,
		Payload:   mustMarshal(t, domain.MessagePayload{Message: msg}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	pub.EXPECT().PublishNewQuestion(ctx, msg).Return(nil)
	repo.EXPECT().MarkPublished(ctx, eventID).Return(nil)

	relay.process(ctx)
}

func TestProcess_FollowUp(t *testing.T) {
	relay, repo, pub, _, _ := newRelay(t)
	ctx := context.Background()

	msg := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventFollowUp,
		Payload:   mustMarshal(t, domain.MessagePayload{Message: msg}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	pub.EXPECT().PublishFollowUp(ctx, msg.RoomID, msg).Return(nil)
	repo.EXPECT().MarkPublished(ctx, eventID).Return(nil)

	relay.process(ctx)
}

func TestProcess_AIReply(t *testing.T) {
	relay, repo, pub, _, _ := newRelay(t)
	ctx := context.Background()

	humanID := uuid.New()
	msg := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventAIReply,
		Payload:   mustMarshal(t, domain.MessagePayload{Message: msg, HumanID: humanID}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	pub.EXPECT().PublishAIReply(ctx, humanID, msg).Return(nil)
	repo.EXPECT().MarkPublished(ctx, eventID).Return(nil)

	relay.process(ctx)
}

func TestProcess_TagsSync(t *testing.T) {
	relay, repo, _, syncer, _ := newRelay(t)
	ctx := context.Background()

	userID := uuid.New()
	tags := []string{"go", "python"}
	oldTags := []string{"go"}
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventTagsSync,
		Payload:   mustMarshal(t, domain.TagSyncPayload{UserID: userID, Tags: tags, OldTags: oldTags}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	syncer.EXPECT().SyncAIQueue(ctx, userID, tags, oldTags).Return(nil)
	repo.EXPECT().MarkPublished(ctx, eventID).Return(nil)

	relay.process(ctx)
}

func TestProcess_RoomClaimed(t *testing.T) {
	relay, repo, _, _, binder := newRelay(t)
	ctx := context.Background()

	roomID := uuid.New()
	aiID := uuid.New()
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventRoomClaimed,
		Payload:   mustMarshal(t, domain.RoomClaimedPayload{RoomID: roomID, AiID: aiID}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	binder.EXPECT().BindRoomToAI(ctx, roomID, aiID).Return(nil)
	repo.EXPECT().MarkPublished(ctx, eventID).Return(nil)

	relay.process(ctx)
}

func TestProcess_DispatchError_MarksAsFailed(t *testing.T) {
	relay, repo, pub, _, _ := newRelay(t)
	ctx := context.Background()

	msg := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventNewQuestion,
		Payload:   mustMarshal(t, domain.MessagePayload{Message: msg}),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	pub.EXPECT().PublishNewQuestion(ctx, msg).Return(errors.New("broker down"))
	repo.EXPECT().MarkFailed(ctx, eventID, "PublishNewQuestion: broker down").Return(nil)

	relay.process(ctx)
}

func TestProcess_UnknownEventType_MarksAsFailed(t *testing.T) {
	relay, repo, _, _, _ := newRelay(t)
	ctx := context.Background()

	eventID := uuid.New()
	event := &domain.OutboxEvent{
		ID:        eventID,
		EventType: domain.EventType("unknown"),
		Payload:   json.RawMessage(`{}`),
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return([]*domain.OutboxEvent{event}, nil)
	repo.EXPECT().MarkFailed(ctx, eventID, "unknown event type: unknown").Return(nil)

	relay.process(ctx)
}

func TestProcess_MultipleEvents(t *testing.T) {
	relay, repo, pub, _, _ := newRelay(t)
	ctx := context.Background()

	msg1 := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	msg2 := &domain.Message{ID: uuid.New(), RoomID: uuid.New()}
	id1, id2 := uuid.New(), uuid.New()
	events := []*domain.OutboxEvent{
		{ID: id1, EventType: domain.EventNewQuestion, Payload: mustMarshal(t, domain.MessagePayload{Message: msg1})},
		{ID: id2, EventType: domain.EventNewQuestion, Payload: mustMarshal(t, domain.MessagePayload{Message: msg2})},
	}

	repo.EXPECT().FetchPending(ctx, fetchLimit).Return(events, nil)
	pub.EXPECT().PublishNewQuestion(ctx, msg1).Return(nil)
	repo.EXPECT().MarkPublished(ctx, id1).Return(nil)
	pub.EXPECT().PublishNewQuestion(ctx, msg2).Return(nil)
	repo.EXPECT().MarkPublished(ctx, id2).Return(nil)

	relay.process(ctx)
}
