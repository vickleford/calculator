package workqueue_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vickleford/calculator/internal/workqueue"
)

type consumerHandlerFunc func(context.Context, []byte) error

func (h consumerHandlerFunc) Handle(ctx context.Context, payload []byte) error {
	return h(ctx, payload)
}

func TestIntegration_Workqueue_PublishAndConsumeJSON(t *testing.T) {
	rmqURL := os.Getenv("RMQ_URL")
	if rmqURL == "" {
		t.Skip(`set RMQ_URL to run this test, e.g. RMQ_URL="amqp://guest:guest@localhost:5672/"`)
	}

	const queueName = "integration"

	type message struct {
		Hello string `json:"hello"`
		ID    string `json:"id"`
	}

	id := uuid.New().String()

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		t.Fatalf("cannot reach rmq server: %s", err)
	}
	defer conn.Close()

	producer := workqueue.NewProducer(conn, workqueue.WithQueueName(queueName))
	defer producer.Close()

	msg := message{Hello: "world", ID: id}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := producer.PublishJSON(ctx, msg); err != nil {
		t.Errorf("publishing JSON yielded unexpected error: %s", err)
	}

	receivedDesiredMessage := make(chan struct{})
	h := consumerHandlerFunc(func(ctx context.Context, payload []byte) error {
		var msg message
		if err := json.Unmarshal(payload, &msg); err != nil {
			t.Errorf("unexpected error ummarshaling JSON: %s", err)
		}
		if msg.Hello == "world" && msg.ID == id {
			receivedDesiredMessage <- struct{}{}
		}
		return nil
	})

	consumerFinishedUnexpectedly := make(chan error)
	consumer := workqueue.NewConsumer(conn, h)
	go func() {
		consumerFinishedUnexpectedly <- consumer.Start(ctx)
	}()

	select {
	case <-receivedDesiredMessage:
	case err := <-consumerFinishedUnexpectedly:
		t.Log(err)
		t.Errorf("unexpected consumer stop: %s", err)
	case <-time.After(10 * time.Second):
		t.Errorf("never received desired message")
	}
}
