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
