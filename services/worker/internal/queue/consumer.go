package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"hafiztri123/worker-link-shortener/internal/queue/metadata"
	"hpj/hv1-link-shortener/shared/models"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MaxRetries = 3
)

type Consumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    queueLabel   string
	metadataRepository *metadata.Repository
}

func NewConsumer(addr, queueLabel string, metadataRepository *metadata.Repository) (*Consumer, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		slog.Error("failed to dial rabbit mq server", "err", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to established channel in rabbit mq connection", "err", err)
		conn.Close()
		return nil, err
	}

	if err := setupDeadLetterQueue(ch, queueLabel); err != nil {
		slog.Error("failed to create dead letter exchange", "error", err)
		conn.Close()
		ch.Close()
		return nil, err
	}

	if err := setupRetryQueue(ch, queueLabel); err != nil {
		slog.Error("failed to create retry queue", "error", err)
		conn.Close()
		ch.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueLabel,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange": queueLabel + ".dlx",
			"x-dead-letter-routing-key": queueLabel + ".dlq",
		},
	)

	 if err != nil {
        ch.Close()
        conn.Close()
		slog.Error("failed to queue declare with dead letter mechanism", "error", err)
        return nil, err
    }

	err = ch.Qos(1, 0, false)

	if err != nil {
        ch.Close()
        conn.Close()
		slog.Error("failed to set quality of service", "error", err)
        return nil, err
    }
	


	return &Consumer{
		conn: conn,
		channel: ch,
		queueLabel: queueLabel,
		metadataRepository: metadataRepository,
	}, nil


}

func (c *Consumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queueLabel,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to consume", "error", err)
		return err
	}

	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		
		case msg, ok := <-msgs:
			if !ok {
				slog.Error("error in getting value from channel")
				return fmt.Errorf("channel closed")
			}

			c.handleMetadataMessage(msg)
		}
	}
}

func (c *Consumer) handleMetadataMessage(msg amqp.Delivery) {
	retryCount := getRetryCount(msg.Headers)
	var data *models.Click

	err := json.Unmarshal(msg.Body, &data)

	if err != nil {
		slog.Error("failed to handle data", "err", err)
		return 
	}

	contextTimeout, cancel := context.WithTimeout(context.Background(), 5 *time.Second)
	defer cancel()

	if err := c.metadataRepository.InsertMetadata(contextTimeout, data); err != nil {
		if retryCount >= MaxRetries {
			msg.Nack(false, false)
			return
		}

		if err := c.sendToRetryQueue(msg, retryCount+1); err != nil {
			msg.Nack(false, true)
			return
		}

		msg.Ack(false)
		return
	}

	msg.Ack(false)

}

func (c *Consumer) sendToRetryQueue(msg amqp.Delivery, retryCount int32) error {
	retryQueueLabel := c.queueLabel + ".retry"
	retryExchangeLabel := c.queueLabel + ".retry.exchange"

	headers := make(amqp.Table)
	if msg.Headers != nil {
		headers = msg.Headers
	}
	headers["x-retry-count"] = retryCount

	return c.channel.Publish(
		retryExchangeLabel,
		retryQueueLabel,
		false,
		false,
		amqp.Publishing{
			ContentType: msg.ContentType,
			Body: msg.Body,
			DeliveryMode: amqp.Persistent,
			Headers: headers,
		},
	)
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	return nil
}
