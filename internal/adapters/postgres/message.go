package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (pg *postgres) Save(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
	var message domain.Message
	if err := pg.DB.GetContext(
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
	return &message, nil
}

func (pg *postgres) GetLastMessages(
	ctx context.Context,
	roomID string,
	limit int,
) ([]*domain.Message, error) {
	var messages []*domain.Message

	err := pg.DB.SelectContext(ctx, &messages, getLastMessagesQuery, roomID, limit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get last messages: %w", err)
	}
	return messages, nil
}

func (pg *postgres) ClaimMessage(ctx context.Context, messageID uuid.UUID, aiID uuid.UUID) error {
	tx, err := pg.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // no-op if tx already committed
	}()

	var currentAssignee *uuid.UUID
	if err := tx.GetContext(ctx, &currentAssignee, getMessageAssigneeQuery, messageID); err != nil {
		return fmt.Errorf("get message assignee: %w", err)
	}

	if currentAssignee != nil {
		return errors.New("message already claimed by another AI")
	}

	if _, err := tx.ExecContext(ctx, claimMessageQuery, aiID, messageID); err != nil {
		return fmt.Errorf("claim message: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
