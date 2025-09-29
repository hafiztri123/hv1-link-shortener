package queue

import amqp "github.com/rabbitmq/amqp091-go"

func getRetryCount(headers amqp.Table) int32 {
	if headers == nil {
		return 0
	}

	if count, ok := headers["x-retry-count"]; ok {
		switch v := count.(type) {
			case int32: 
				return v
		case int:
				return int32(v)
		case int64:
			return int32(v)
		}
	}

	return 0
}