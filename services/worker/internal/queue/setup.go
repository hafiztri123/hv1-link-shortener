package queue

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

func setupDeadLetterQueue(ch *amqp.Channel, queueLabel string) error {
	dlxName := queueLabel + ".dlx"
	err := ch.ExchangeDeclare(
		dlxName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("error exchange declare for dead letter exchange", "error", err )
		return err
	}

	dlqName := queueLabel + ".dlq"
	_, err = ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("error queue declare for dead letter queue", "error", err)
		return err
	}

	err = ch.QueueBind(
		dlqName,
		dlqName,
		dlxName,
		false,
		nil,
	)

	if err != nil {
		slog.Error("error queue bind declare", "error", err)
		return err
	}

	return nil
}

func setupRetryQueue(ch *amqp.Channel, queueLabel string) error {
	retryQueueLabel := queueLabel + ".retry"
	retryExchangeLabel := queueLabel + ".retry.exchange"


	err := ch.ExchangeDeclare(
		retryExchangeLabel,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		slog.Error("failed to exchange decrlare for retry queue")
		return err
	}

	_, err = ch.QueueDeclare(
		retryQueueLabel,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl" : 30000,
			"x-dead-letter-exchange": "",
			"x-dead-letter-routing-key": queueLabel,
		},
	)

	if err != nil {
		slog.Error("failed to queue declare for retry queue")
		return err
	}

	err = ch.QueueBind(
		retryQueueLabel,
		retryQueueLabel,
		retryExchangeLabel,
		false,
		nil,
	)

	return nil
}