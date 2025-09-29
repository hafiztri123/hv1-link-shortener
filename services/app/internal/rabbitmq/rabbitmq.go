package rabbitmq

import (

	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)


type RabbitMQ struct {
	conn *amqp.Connection
	channel *amqp.Channel
	queueLabel string
}


func NewRabbitMQ(addr, queueLabel string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		slog.Error("error in dialing rabbit mq", "err", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		slog.Error("failed to established channel in connection", "err", err)
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueLabel,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to declare queue")
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn: conn,
		channel: ch,
		queueLabel: queueLabel,
	}, nil
}



func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}

	if r.conn != nil {
		r.conn.Close()
	}

	return nil
}

func (r *RabbitMQ) HealthCheck() error {
	return r.channel.ExchangeDeclarePassive(
		"amq.direct",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
}