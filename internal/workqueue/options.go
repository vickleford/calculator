package workqueue

type AMQP091Option[T Producer | AMQP091Consumer] func(*T)

func WithQueueName[T Producer | AMQP091Consumer](name string) AMQP091Option[T] {
	return func(t *T) {
		switch concrete := any(t).(type) {
		case *Producer:
			concrete.desiredQueueName = name
		case *AMQP091Consumer:
			concrete.desiredQueueName = name
		default:
			panic("unsupported type")
		}
	}
}

func WithDurableQueue[T Producer | AMQP091Consumer]() AMQP091Option[T] {
	return func(t *T) {
		switch concrete := any(t).(type) {
		case *Producer:
			concrete.durable = true
		case *AMQP091Consumer:
			concrete.durable = true
		default:
			panic("unsupported type")
		}
	}
}
