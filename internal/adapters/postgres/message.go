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

func (pg *postgres) SaveWithOutbox(ctx context.Context, msg *domain.Message, eventType domain.EventType, recipientID uuid.UUID) (*domain.Message, error) {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("tx begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

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
		Message:     &message,
		RecipientID: recipientID,
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

func (pg *postgres) GetLastMessage(ctx context.Context, roomID uuid.UUID) (*domain.Message, error) {
	var message domain.Message
	err := pg.DB.GetContext(ctx, &message, getLastMessageQuery, roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNoMessages
	}
	if err != nil {
		return nil, fmt.Errorf("get last message: %w", err)
	}
	return &message, nil
}
