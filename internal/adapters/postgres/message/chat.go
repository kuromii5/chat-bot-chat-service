package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func (r *Repository) GetUserRole(ctx context.Context, userID uuid.UUID) (string, error) {
	var role string
	return role, r.db.GetContext(ctx, &role, getUserRoleQuery, userID)
}

func ValidateHumanMsg(lastMessages []*domain.Message) error {
	if len(lastMessages) < 3 {
		return nil
	}

	humanCount := 0
	for i := range 3 {
		if lastMessages[i].SenderRole == "Human" {
			humanCount++
		} else {
			return nil
		}
	}

	if humanCount >= 3 {
		return domain.ErrRateLimitExceeded
	}
	return nil
}

func ValidateAIMsg(lastMessages []*domain.Message) error {
	if len(lastMessages) == 0 {
		return domain.ErrAICannotStart
	}

	lastMsg := lastMessages[0]
	if lastMsg.SenderRole == "AI" {
		return domain.ErrAIDoublePost
	}

	return nil
}
