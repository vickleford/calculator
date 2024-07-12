package workqueue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrChannelClosed = errors.New("channel was closed")

type AMQP091Consumer struct {
	conn amqpConnection
	msgs <-chan amqp.Delivery
}

func NewConsumer(conn amqpConnection) *AMQP091Consumer {
	return &AMQP091Consumer{conn: conn}
}

func (c *AMQP091Consumer) Start(ctx context.Context) error {
	recvCh, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("error opening channel: %w", err)
	}

	defer recvCh.Close()

	q, err := recvCh.QueueDeclare(
		"desiredQueueName", // name
		false,              // durable
		false,              // autodelete
		false,              // exclusive
		false,              // noWait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("error declaring queue: %w", err)
	}

	msgs, err := recvCh.Consume(
		q.Name,
		"",    // consumer identifier
		false, // autoAck
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("error establishing message delivery channel: %w", err)
	}

	c.msgs = msgs

	return nil
}

func (c *AMQP091Consumer) Messages() <-chan amqp.Delivery {
	return c.msgs
}

type consumer interface {
	Start(context.Context) error
	Messages() <-chan amqp.Delivery
}

type handler interface {
	Handle(context.Context, []byte) error
}

type FibOfConsumer struct {
	consumer consumer
	strategy handler
}

// FibOfConsumer returns a new FibOfConsumer to produce *FibonacciOfJob messages.
// The consumer must be started separately.
func NewFibOfConsumer(consumer consumer, strategy handler) *FibOfConsumer {
	return &FibOfConsumer{consumer: consumer, strategy: strategy}
}

func (f *FibOfConsumer) NextFibOfJob(ctx context.Context) (*FibonacciOfJob, error) {
	var delivery amqp.Delivery

	select {
	case delivery = <-f.consumer.Messages():
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if len(delivery.Body) == 0 && delivery.MessageCount == 0 {
		return nil, ErrChannelClosed
	}

	var job *FibonacciOfJob

	if err := json.Unmarshal(delivery.Body, job); err != nil {
		if rejectErr := delivery.Reject(true); rejectErr != nil {
			return nil, NewAcknowledgementError(AcknowledgementErrorOpReject, rejectErr, err)
		}
		return nil, fmt.Errorf("error unmarshaling message: %w", err)
	}

	if err := delivery.Ack(false); err != nil {
		return nil, NewAcknowledgementError(AcknowledgementErrorOpAck, err, nil)
	}

	return job, nil
}
