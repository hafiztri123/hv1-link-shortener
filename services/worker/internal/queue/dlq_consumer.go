package queue

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DLQConsumer struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	queueLabel string
}

func NewDLQConsumer(addr, queueLabel string) (*DLQConsumer, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		slog.Error("error dialing dlq consumer", "error", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to established channel in dlq connection", "err", err)
		conn.Close()
		return nil, err
	}

	return &DLQConsumer{
		conn:       conn,
		ch:         ch,
		queueLabel: queueLabel + ".dlq",
	}, nil
}

func (c *DLQConsumer) StartConsuming(ctx context.Context) error {
	msgs, err := c.ch.Consume(
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
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-msgs:
			if !ok {
				slog.Error("error in getting value from channel")
				return fmt.Errorf("channel closed")
			}

			//TODO

			msg.Ack(false)
		}
	}
}

func (c *DLQConsumer) Close() error {
	if c.ch != nil {
		c.ch.Close()
	}

	if c.conn != nil {
		c.conn.Close()
	}

	return nil
}
