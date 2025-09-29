package rabbitmq

import (
	"context"
	"hpj/hv1-link-shortener/shared/models"
	"testing"
	"time"
)

func TestRabbitMQ_NewRabbitMQ_InvalidAddr(t *testing.T) {
	_, err := NewRabbitMQ("invalid://addr", "test-queue")
	if err == nil {
		t.Error("Expected error for invalid RabbitMQ address")
	}
}

func TestRabbitMQ_Close_WithNilConnections(t *testing.T) {
	rmq := &RabbitMQ{
		conn:    nil,
		channel: nil,
	}

	err := rmq.Close()
	if err != nil {
		t.Errorf("Expected no error when closing nil connections, got: %v", err)
	}
}

func TestRabbitMQ_HealthCheck_WithNilChannel(t *testing.T) {
	rmq := &RabbitMQ{
		conn:    nil,
		channel: nil,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when health checking with nil channel")
		}
	}()

	_ = rmq.HealthCheck()
}

func TestRabbitMQ_PublishClickEvent_WithNilChannel(t *testing.T) {
	rmq := &RabbitMQ{
		channel: nil,
	}

	clickEvent := &models.Click{
		Path:      "/test",
		IPAddress: "127.0.0.1",
		Timestamp: time.Now(),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when publishing with nil channel")
		}
	}()

	_ = rmq.PublishClickEvent(context.Background(), clickEvent)
}