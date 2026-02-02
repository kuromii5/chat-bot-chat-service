package rabbitmq

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *RabbitMQ) SetupAIQueue(ctx context.Context, userID uuid.UUID, tags []string) error {
	queueName := fmt.Sprintf("ai_queue_%s", userID.String())

	// 1. Создаем железную (durable) очередь
	_, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable: выживет после рестарта Кролика
		false, // auto-delete: не удалять, когда AI уйдет в оффлайн
		false, // exclusive: другие тоже могут подключиться (если надо)
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 2. Чистим старые биндинги (опционально, зависит от логики,
	// но лучше пересобрать связи заново)
	// ... (тут логика Unbind, если теги изменились)

	// 3. Вяжем теги к очереди
	for _, tag := range tags {
		routingKey := fmt.Sprintf("question.#.%s.#", tag)
		err = r.channel.QueueBind(queueName, routingKey, r.config.Exchange, false, nil)
		if err != nil {
			return err
		}
	}
	return nil
}
