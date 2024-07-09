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

func TestIntegration_Producer_PublishJSON(t *testing.T) {
	rmqURL := os.Getenv("RMQ_URL")
	if rmqURL == "" {
		t.Skip(`set RMQ_URL to run this test, e.g. RMQ_URL="amqp://guest:guest@localhost:5672/"`)
	}

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

	producer := workqueue.NewProducer(conn, workqueue.WithQueueName("integration"))
	defer producer.Close()

	msg := message{Hello: "world", ID: id}

	if err := producer.PublishJSON(context.Background(), msg); err != nil {
		t.Errorf("publishing JSON yielded unexpected error: %s", err)
	}

	recvCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("error setting up receive channel: %s", err)
	}

	defer recvCh.Close()
	q, err := recvCh.QueueDeclare("integration", false, false, false, false, nil)
	if err != nil {
		t.Errorf("error declaring queue: %s", err)
	}

	msgs, err := recvCh.Consume(
		q.Name,
		"",
		false, // autoack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("unable to set up consumer: %s", err)
	}

	receivedDesiredMessage := make(chan struct{})

	go func() {
		for delivery := range msgs {
			var msg message
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				t.Errorf("unexpected error ummarshaling JSON: %s", err)
			}
			if msg.Hello == "world" && msg.ID == id {
				if err := delivery.Ack(false); err != nil {
					t.Errorf("unexpected error acking message: %s", err)
				}
				receivedDesiredMessage <- struct{}{}
			}
		}
	}()

	select {
	case <-receivedDesiredMessage:
	case <-time.After(10 * time.Second):
		t.Errorf("never received desired message")
	}
}
