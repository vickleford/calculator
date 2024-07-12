package workqueue

import (
	"context"
	"errors"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrChannelClosed = errors.New("channel was closed")

type handler interface {
	Handle(context.Context, []byte) error
}

type AMQP091Consumer struct {
	conn     amqpConnection
	strategy handler

	desiredQueueName string
	durable          bool
}

func NewConsumer[T AMQP091Consumer](conn amqpConnection, strategy handler, opts ...AMQP091Option[T]) *T {
	c := new(T)

	concrete, ok := any(c).(*AMQP091Consumer)
	if !ok {
		panic("unsupported consumer type")
	}
	concrete.conn = conn
	concrete.strategy = strategy

	for _, o := range opts {
		o(c)
	}

	return c
}

func (c *AMQP091Consumer) Start(ctx context.Context) error {
	if c.strategy == nil {
		return fmt.Errorf("no strategy provided for handling messages")
	}

	recvCh, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("error opening channel: %w", err)
	}

	defer recvCh.Close()

	q, err := recvCh.QueueDeclare(
		c.desiredQueueName, // name
		c.durable,          // durable
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

	for {
		if err := c.receive(ctx, msgs); errors.Is(err, ErrChannelClosed) {
			return err
		} else if ackErr, ok := err.(*AcknowledgementError); ok {
			return ackErr
		} else if errors.Is(err, context.Canceled) {
			return err
		} else if err != nil {
			log.Printf("error handling message: %s", err)
		}
	}
}

func (c *AMQP091Consumer) receive(ctx context.Context, msgs <-chan amqp.Delivery) error {
	var delivery amqp.Delivery

	select {
	case delivery = <-msgs:
	case <-ctx.Done():
		return ctx.Err()
	}

	if len(delivery.Body) == 0 && delivery.MessageCount == 0 {
		return ErrChannelClosed
	}

	if err := c.strategy.Handle(ctx, delivery.Body); err != nil {
		const alwaysRequeue = true
		if rejectErr := delivery.Reject(alwaysRequeue); rejectErr != nil {
			return NewAcknowledgementError(AcknowledgementErrorOpReject, rejectErr, err)
		}
		return fmt.Errorf("error handling message: %w", err)
	}

	const ackMultipleMessages = true
	if err := delivery.Ack(!ackMultipleMessages); err != nil {
		return NewAcknowledgementError(AcknowledgementErrorOpAck, err, nil)
	}

	return nil
}
