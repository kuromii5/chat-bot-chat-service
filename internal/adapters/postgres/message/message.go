package message

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
	"github.com/lib/pq"
)

func (r *Repository) Save(ctx context.Context, msg *domain.Message) (*domain.Message, error) {
	var message domain.Message
	return &message, r.db.GetContext(ctx, &message, saveMessageQuery, msg.SenderID, msg.SenderRole, msg.RoomID, msg.Content, pq.Array(msg.Tags))
}

func (r *Repository) GetLastMessages(ctx context.Context, roomID string, limit int) ([]*domain.Message, error) {
	var messages []*domain.Message

	err := r.db.SelectContext(ctx, &messages, getLastMessagesQuery, roomID, limit)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return messages, nil
}

func (r *Repository) ClaimMessage(ctx context.Context, messageID uuid.UUID, aiID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentAssignee *uuid.UUID
	if err := tx.GetContext(ctx, &currentAssignee, getMessageAssigneeQuery, messageID); err != nil {
		return err
	}

	if currentAssignee != nil {
		return errors.New("message already claimed by another AI")
	}

	_, err = tx.ExecContext(ctx, claimMessageQuery, aiID, messageID)

	return tx.Commit()
}
