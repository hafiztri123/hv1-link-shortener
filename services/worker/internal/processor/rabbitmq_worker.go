package processor

import (
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQWorker struct {
	conn        *amqp.Connection
	queueName   string
	workerCount int
	batchSize   int
	delay       time.Duration
	done        chan struct{}
	wg          sync.WaitGroup
}

func NewRabbitMQWorker(rabbitAddr, queueName string, workerCount, batchSize int, delay time.Duration) (*RabbitMQWorker, error) {
	conn, err := amqp.Dial(rabbitAddr)
	if err != nil {
		slog.Error("failed to dial rabbit mq server", "err", err)
		return nil, err
	}

	return &RabbitMQWorker{
		conn:        conn,
		queueName:   queueName,
		workerCount: workerCount,
		batchSize:   batchSize,
		delay:       delay,
		done:        make(chan struct{}),
	}, nil
}

// func (r *RabbitMQWorker) Start (ctx context.Context) error {
// 	slog.Info("Starting rabbit mq worker",
// 		"queue" , r.queueName,
// 		"workers", r.workerCount,
// 		"batch_size", r.batchSize,
// 	)
// }

// func (r *RabbitMQWorker) worker(ctx context.Context, workerID int) {
// 	defer r.wg.Done()

// 	ch, err := r.conn.Channel()
// 	if err != nil {
// 		slog.Error("Failed to open channel in rabbit mq server", "err", err)
// 		return
// 	}

// 	defer ch.Close()

// 	err = ch.Qos(
// 		r.batchSize,
// 		0,
// 		false,
// 	)

// 	if err != nil {
// 		slog.Error("failed to set QoS", "err", err)
// 		return
// 	}

// 	messages, err := ch.Consume(
// 		r.queueName,
// 		fmt.Sprintf("worker-%d", workerID),
// 		false,
// 		false,
// 		false,
// 		false,
// 		nil,
// 	)

// 	if err != nil {
// 		slog.Error("consume failed", "worker", workerID, "error", err)
// 		return
// 	}

// 	slog.Info("worker started", "worker", workerID)

// 	batch := make([]models.Click, 0, r.batchSize)
// 	deliveries := make([]amqp.Delivery, 0, r.batchSize)
// 	ticker := time.NewTicker(r.delay)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-r.done:
// 			if len(batch) > 0 {
// 				//TODO
// 			}
// 			return

// 		case <-ctx.Done():
// 			return

// 		case delivery, ok := <-messages:
// 			if !ok {
// 				slog.Warn("message channel closed", "worker", workerID)
// 				return
// 			}
// 			//TODO

// 		case <-ticker.C:
// 			//TODO
// 		}

// 	}

// }
