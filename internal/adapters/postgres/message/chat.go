package message

import (
	"github.com/kuromii5/chat-bot-chat-service/internal/domain"
)

func ValidateHumanMsg(lastMessages []*domain.Message) error {
	if len(lastMessages) < 3 {
		return nil
	}

	humanCount := 0
	for i := range 3 {
		if lastMessages[i].SenderRole == domain.Human {
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
	if lastMsg.SenderRole == domain.AI {
		return domain.ErrAIDoublePost
	}

	return nil
}
