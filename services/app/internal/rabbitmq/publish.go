package rabbitmq

import (
	"context"
	"encoding/json"
	"hpj/hv1-link-shortener/shared/models"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQ) PublishClickEvent(ctx context.Context, clickEvent *models.Click) error {
	body, err := json.Marshal(clickEvent)
	if err != nil {
		slog.Error("failed to marshal click event", "Err", err)
		return err
	}

	err = r.channel.PublishWithContext(
		ctx,
		"",
		r.queueLabel,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		slog.Error("failed to publish click event", "error", err)
		return err
	}

	return nil
}
