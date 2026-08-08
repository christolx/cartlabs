//go:build integration

package messaging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type integrationPublisher struct{ err error }

func (p integrationPublisher) Publish(context.Context, string, []byte) error { return p.err }

func TestOutboxBackoffDeadLetterAndReplay(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, `UPDATE outbox_events SET published_at=now() WHERE published_at IS NULL`); err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	now := time.Now().UTC()
	if _, err := pool.Exec(ctx, `INSERT INTO outbox_events (id,aggregate_type,aggregate_id,event_type,payload,occurred_at,next_attempt_at)
		VALUES ($1,'purchase',$2,'purchase.paid','{"purchaseId":"01989f00-0000-7000-8000-000000000701"}'::jsonb,$3,$3)`, id, uuid.NewString(), now); err != nil {
		t.Fatal(err)
	}
	outbox := NewOutbox(pool)
	failure := errors.New("broker unavailable")
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		claimed, err := outbox.RelayOne(ctx, integrationPublisher{err: failure}, now.Add(time.Duration(attempt)*2*time.Minute))
		if !claimed || !errors.Is(err, failure) {
			t.Fatalf("attempt=%d claimed=%v err=%v", attempt, claimed, err)
		}
	}
	var attempts int
	var deadLetteredAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT attempts,dead_lettered_at FROM outbox_events WHERE id=$1`, id).Scan(&attempts, &deadLetteredAt); err != nil {
		t.Fatal(err)
	}
	if attempts != maxAttempts || deadLetteredAt == nil {
		t.Fatalf("attempts=%d deadLetteredAt=%v", attempts, deadLetteredAt)
	}
	replayed, err := outbox.ReplayDeadLetters(ctx, 10, now.Add(20*time.Minute))
	if err != nil || replayed < 1 {
		t.Fatalf("replayed=%d err=%v", replayed, err)
	}
	claimed, err := outbox.RelayOne(ctx, integrationPublisher{}, now.Add(21*time.Minute))
	if err != nil || !claimed {
		t.Fatalf("successful replay claimed=%v err=%v", claimed, err)
	}
	var publishedAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT published_at FROM outbox_events WHERE id=$1`, id).Scan(&publishedAt); err != nil {
		t.Fatal(err)
	}
	if publishedAt == nil {
		t.Fatal("replayed outbox event was not published")
	}
}

type failingHandler struct{}

func (failingHandler) Handle(context.Context, []byte) error {
	return errors.New("transient handler failure")
}

func TestRabbitRetriesThenDeadLetters(t *testing.T) {
	rabbitURL := os.Getenv("INTEGRATION_RABBITMQ_URL")
	if rabbitURL == "" {
		t.Skip("INTEGRATION_RABBITMQ_URL is not set")
	}
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	queue := "cartlabs.notifications.integration"
	consumer, err := NewNotificationConsumer(connection, DefaultExchange, queue)
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Close()
	inspection, err := connection.Channel()
	if err != nil {
		t.Fatal(err)
	}
	defer inspection.Close()
	defer func() {
		_, _ = inspection.QueuePurge(DefaultQueue, false)
		_, _ = inspection.QueuePurge(RetryQueue, false)
		_, _ = inspection.QueuePurge(DeadQueue, false)
	}()
	if _, err := inspection.QueuePurge(DeadQueue, false); err != nil {
		t.Fatal(err)
	}
	if _, err := inspection.QueuePurge(queue, false); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	consumerDone := make(chan error, 1)
	go func() {
		consumerDone <- consumer.Run(ctx, failingHandler{}, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	}()
	publisher, err := NewRabbitPublisher(connection, DefaultExchange)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Close()
	eventID := uuid.NewString()
	body := []byte(`{"id":"` + eventID + `","type":"purchase.paid","occurredAt":"` + time.Now().UTC().Format(time.RFC3339Nano) + `","data":{}}`)
	if err := publisher.Publish(ctx, "purchase.paid", body); err != nil {
		t.Fatal(err)
	}
	for ctx.Err() == nil {
		delivery, ok, err := inspection.Get(DeadQueue, false)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			if string(delivery.Body) != string(body) || retryCount(delivery.Headers) != maxDeliveries {
				t.Fatalf("dead letter body=%s retries=%d", delivery.Body, retryCount(delivery.Headers))
			}
			if err := delivery.Nack(false, true); err != nil {
				t.Fatal(err)
			}
			cancel()
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatal("notification was not dead-lettered")
	}
	if err := <-consumerDone; err != nil {
		t.Fatalf("consumer shutdown: %v", err)
	}
	if err := consumer.Close(); err != nil {
		t.Fatal(err)
	}
	replayer, err := NewDeadLetterReplayer(connection, DefaultExchange)
	if err != nil {
		t.Fatal(err)
	}
	defer replayer.Close()
	replayCtx, replayCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer replayCancel()
	replayed, err := replayer.Replay(replayCtx, 1)
	if err != nil || replayed != 1 {
		t.Fatalf("replayed=%d err=%v", replayed, err)
	}
	replayedDelivery, ok, err := inspection.Get(queue, true)
	if err != nil || !ok || string(replayedDelivery.Body) != string(body) || retryCount(replayedDelivery.Headers) != 0 {
		t.Fatalf("replayed delivery ok=%v err=%v body=%s retries=%d", ok, err, replayedDelivery.Body, retryCount(replayedDelivery.Headers))
	}
}
