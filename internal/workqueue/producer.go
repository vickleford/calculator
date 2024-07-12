package workqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpConnection interface {
	Channel() (*amqp.Channel, error)
}

// Producer is capable of publishing to a RabbitMQ exchange.
type Producer struct {
	desiredQueueName string
	durable          bool
	deleteWhenUnused bool
	exclusive        bool
	noWait           bool

	exchange  string
	mandatory bool
	immediate bool

	queue   amqp.Queue
	channel *amqp.Channel

	requestReinit chan struct{}
	reinitialize  chan struct{}
	ready         chan struct{}
}

func NewProducer[T Producer](conn amqpConnection, opts ...AMQP091Option[T]) *T {
	p := new(T)

	producer, ok := any(p).(*Producer)
	if !ok {
		panic("unsupported producer type")
	}
	producer.reinitialize = make(chan struct{})
	producer.ready = make(chan struct{})

	for _, o := range opts {
		o(p)
	}

	go func() {
		// TODO: consider a backoff for retries.
		// TODO: protect this routine from leaking.
		for {
			ch, err := conn.Channel()
			if err != nil {
				log.Printf("error creating channel: %s", err)
				continue
				// ???
			}

			q, err := ch.QueueDeclare(
				producer.desiredQueueName,
				producer.durable,
				producer.deleteWhenUnused,
				producer.exclusive,
				producer.noWait,
				nil,
			)
			if err != nil {
				log.Printf("error declaring queue: %s", err)
				ch.Close()
				continue
				// ???
			}

			producer.channel = ch
			producer.queue = q
			producer.ready <- struct{}{}

			<-producer.reinitialize
			ch.Close()
		}
	}()

	<-producer.ready

	// Handle reinitialize requests. Essentially, multiple requests from
	// multiple callers need to be trimmed down to one signal to reinitialize
	// without everything getting backed up.
	go func() {
		var waitingForReady bool
		for {
			<-producer.requestReinit
			// Drain off reinitialize requests from multiple callers until we
			// are ready. This helps multiple callers from stampeding the
			// reinitialization of the channel.
			if !waitingForReady {
				waitingForReady = true
				go func() {
					producer.reinitialize <- struct{}{}
					<-producer.ready
					waitingForReady = false
				}()
			}
		}
	}()

	return p
}

func (p *Producer) requestChannelReinitialization() {
	p.requestReinit <- struct{}{}
}

func (p *Producer) PublishJSON(ctx context.Context, message any) error {
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("unable to marshal message to JSON: %w", err)
	}

	if err := p.channel.PublishWithContext(ctx, p.exchange, p.queue.Name, p.mandatory,
		p.immediate, amqp.Publishing{
			ContentType: "application/json",
			Body:        b,
		}); err != nil {
		p.requestChannelReinitialization()
		return fmt.Errorf("unable to publish message: %s", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.channel.Close()
}
