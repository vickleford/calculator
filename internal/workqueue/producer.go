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

type ProducerOption func(*Producer)

func WithDurableQueue() ProducerOption {
	return func(p *Producer) {
		p.durable = true
	}
}

func WithQueueName(name string) ProducerOption {
	return func(p *Producer) {
		p.desiredQueueName = name
	}
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

func NewProducer(conn amqpConnection, opts ...ProducerOption) *Producer {
	p := &Producer{
		reinitialize: make(chan struct{}),
		ready:        make(chan struct{}),
	}

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

			q, err := ch.QueueDeclare(p.desiredQueueName, p.durable, p.deleteWhenUnused,
				p.exclusive, p.noWait, nil)
			if err != nil {
				log.Printf("error declaring queue: %s", err)
				ch.Close()
				continue
				// ???
			}

			p.channel = ch
			p.queue = q
			p.ready <- struct{}{}

			<-p.reinitialize
			ch.Close()
		}
	}()

	<-p.ready

	// Handle reinitialize requests. Essentially, multiple requests from
	// multiple callers need to be trimmed down to one signal to reinitialize
	// without everything getting backed up.
	go func() {
		var waitingForReady bool
		for {
			<-p.requestReinit
			// Drain off reinitialize requests from multiple callers until we
			// are ready. This helps multiple callers from stampeding the
			// reinitialization of the channel.
			if !waitingForReady {
				waitingForReady = true
				go func() {
					p.reinitialize <- struct{}{}
					<-p.ready
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
