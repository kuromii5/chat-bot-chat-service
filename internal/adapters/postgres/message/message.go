package message

import (
	"context"
	"database/sql"

	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (r *Repository) Save(ctx context.Context, msg *domain.Message) error {
	_, err := r.db.NamedExecContext(ctx, saveMessageQuery, msg)
	return err
}

func (r *Repository) GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error) {
	var messages []*domain.Message

	err := r.db.SelectContext(ctx, &messages, getLastMessagesQuery, roomID, limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return messages, nil
}
