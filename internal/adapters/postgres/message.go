package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (pg *postgres) SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, humanID uuid.UUID) (*domain.Message, error) {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tx begin: %w", err)
	}
	defer tx.Rollback()

	var message domain.Message
	if err := tx.GetContext(
		ctx,
		&message,
		saveMessageQuery,
		msg.SenderID,
		msg.SenderRole,
		msg.RoomID,
		msg.Content,
		pq.Array(msg.Tags),
	); err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	payload, err := json.Marshal(domain.MessagePayload{
		Message: &message,
		HumanID: humanID,
	})
	if err != nil {
		return nil, fmt.Errorf("payload marshal: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		saveOutboxEventQuery,
		"message",
		message.ID,
		eventType,
		payload,
	); err != nil {
		return nil, fmt.Errorf("save message to outbox: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("tx commit: %w", err)
	}
	return &message, nil
}

func (pg *postgres) GetLastMessages(
	ctx context.Context,
	roomID uuid.UUID,
	limit int,
) ([]*domain.Message, error) {
	var messages []*domain.Message
	err := pg.DB.SelectContext(ctx, &messages, getLastMessagesQuery, roomID, limit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get last messages: %w", err)
	}
	return messages, nil
}
