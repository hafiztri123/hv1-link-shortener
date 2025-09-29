package queue

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MaxRetries = 3
)

type Consumer struct {
    conn    *amqp.Connection
    channel *amqp.Channel
    queueLabel   string
}

func NewConsumer(addr, queueLabel string) (*Consumer, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		slog.Error("failed to dial rabbit mq server", "err", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to established channel in rabbit mq connection", "err", err)
		return nil, err
	}


}

func (c *Consumer) StartConsuming(ctx context.Context, handler func([]byte) error) error {
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
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery, handler func([]byte) error){
	retryCount := getRetryCount(msg.Headers)

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
