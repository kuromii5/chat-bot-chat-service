package rabbitmq

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *RabbitMQ) SyncAIQueue(ctx context.Context, userID uuid.UUID, tags, oldTags []string) error {
	queueName := fmt.Sprintf("ai_queue_%s", userID.String())

	if _, err := r.channel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
		return err
	}

	for _, tag := range oldTags {
		routingKey := fmt.Sprintf("question.#.%s.#", tag)
		if err := r.channel.QueueUnbind(queueName, routingKey, r.config.Exchange, nil); err != nil {
			return err
		}
	}

	for _, tag := range tags {
		routingKey := fmt.Sprintf("question.#.%s.#", tag)
		if err := r.channel.QueueBind(queueName, routingKey, r.config.Exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}
