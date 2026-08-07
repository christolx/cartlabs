package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultExchange = "cartlabs.events"
	DefaultQueue    = "cartlabs.notifications"
)

type Event struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	OccurredAt time.Time       `json:"occurredAt"`
	Data       json.RawMessage `json:"data"`
}

type Publisher interface {
	Publish(context.Context, string, []byte) error
}

type Outbox struct{ pool *pgxpool.Pool }

func NewOutbox(pool *pgxpool.Pool) *Outbox { return &Outbox{pool: pool} }

func (o *Outbox) RelayOne(ctx context.Context, publisher Publisher, now time.Time) (bool, error) {
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin outbox relay: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var event Event
	var payload []byte
	err = tx.QueryRow(ctx, `
		SELECT id::text,event_type,occurred_at,payload FROM outbox_events
		WHERE published_at IS NULL ORDER BY occurred_at,id FOR UPDATE SKIP LOCKED LIMIT 1`).
		Scan(&event.ID, &event.Type, &event.OccurredAt, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim outbox event: %w", err)
	}
	event.Data = json.RawMessage(payload)
	body, err := json.Marshal(event)
	if err != nil {
		return true, fmt.Errorf("encode outbox event: %w", err)
	}
	if err := publisher.Publish(ctx, event.Type, body); err != nil {
		message := err.Error()
		if len(message) > 1000 {
			message = message[:1000]
		}
		if _, updateErr := tx.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1,last_error=$2 WHERE id=$1`, event.ID, message); updateErr != nil {
			return true, fmt.Errorf("record outbox failure: %w", updateErr)
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return true, fmt.Errorf("commit outbox failure: %w", commitErr)
		}
		return true, fmt.Errorf("publish outbox event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE outbox_events SET published_at=$2,attempts=attempts+1,last_error='' WHERE id=$1`, event.ID, now); err != nil {
		return true, fmt.Errorf("mark outbox event published: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return true, fmt.Errorf("commit outbox relay: %w", err)
	}
	return true, nil
}

type RabbitPublisher struct {
	channel       *amqp.Channel
	confirmations <-chan amqp.Confirmation
	exchange      string
	mu            sync.Mutex
}

func NewRabbitPublisher(connection *amqp.Connection, exchange string) (*RabbitPublisher, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open publisher channel: %w", err)
	}
	if err := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare event exchange: %w", err)
	}
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("enable publisher confirms: %w", err)
	}
	return &RabbitPublisher{channel: channel, confirmations: channel.NotifyPublish(make(chan amqp.Confirmation, 1)), exchange: exchange}, nil
}

func (p *RabbitPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.channel.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json", DeliveryMode: amqp.Persistent, Timestamp: time.Now().UTC(), Body: body,
	}); err != nil {
		return err
	}
	select {
	case confirmation, ok := <-p.confirmations:
		if !ok || !confirmation.Ack {
			return fmt.Errorf("broker did not confirm publish")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *RabbitPublisher) Close() error { return p.channel.Close() }

type Consumer struct {
	channel  *amqp.Channel
	messages <-chan amqp.Delivery
}

func NewNotificationConsumer(connection *amqp.Connection, exchange, queue string) (*Consumer, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open consumer channel: %w", err)
	}
	closeWith := func(err error) (*Consumer, error) {
		_ = channel.Close()
		return nil, err
	}
	if err := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare consumer exchange: %w", err))
	}
	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare notification queue: %w", err))
	}
	for _, routingKey := range []string{"purchase.#", "order.#"} {
		if err := channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
			return closeWith(fmt.Errorf("bind notification queue: %w", err))
		}
	}
	if err := channel.Qos(10, 0, false); err != nil {
		return closeWith(fmt.Errorf("configure consumer QoS: %w", err))
	}
	messages, err := channel.Consume(queue, "cartlabs-worker", false, false, false, false, nil)
	if err != nil {
		return closeWith(fmt.Errorf("consume notification queue: %w", err))
	}
	return &Consumer{channel: channel, messages: messages}, nil
}

type Handler interface {
	Handle(context.Context, []byte) error
}

func (c *Consumer) Run(ctx context.Context, handler Handler, logger *slog.Logger) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-c.messages:
			if !ok {
				return fmt.Errorf("notification delivery channel closed")
			}
			err := handler.Handle(ctx, delivery.Body)
			if err == nil {
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack notification event: %w", ackErr)
				}
				continue
			}
			permanent := errors.Is(err, domain.ErrInvalid)
			logger.ErrorContext(ctx, "notification event failed", "error", err, "requeue", !permanent)
			if nackErr := delivery.Nack(false, !permanent); nackErr != nil {
				return fmt.Errorf("nack notification event: %w", nackErr)
			}
		}
	}
}

func (c *Consumer) Close() error { return c.channel.Close() }

type NotificationHandler struct{ pool *pgxpool.Pool }

func NewNotificationHandler(pool *pgxpool.Pool) *NotificationHandler {
	return &NotificationHandler{pool: pool}
}

func (h *NotificationHandler) Handle(ctx context.Context, body []byte) error {
	var event Event
	if err := json.Unmarshal(body, &event); err != nil || event.ID == "" || event.Type == "" || event.OccurredAt.IsZero() {
		return domain.ErrInvalid
	}
	if _, err := uuid.Parse(event.ID); err != nil {
		return domain.ErrInvalid
	}
	var data struct {
		PurchaseID   string   `json:"purchaseId"`
		Reference    string   `json:"reference"`
		Status       string   `json:"status"`
		RecipientIDs []string `json:"recipientIds"`
	}
	if err := json.Unmarshal(event.Data, &data); err != nil || data.PurchaseID == "" || data.Reference == "" || len(data.RecipientIDs) == 0 {
		return domain.ErrInvalid
	}
	title, message, ok := notificationCopy(event.Type, data.Reference)
	if !ok {
		return domain.ErrInvalid
	}
	for _, recipientID := range data.RecipientIDs {
		notificationID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate notification ID: %w", err)
		}
		if _, err := h.pool.Exec(ctx, `
			INSERT INTO notifications (id,user_id,source_event_id,kind,title,body,created_at)
			SELECT $1,$2,$3,$4,$5,$6,$7 WHERE EXISTS (SELECT 1 FROM users WHERE id=$2)
			ON CONFLICT (user_id,source_event_id) DO NOTHING`, notificationID.String(), recipientID, event.ID,
			event.Type, title, message, event.OccurredAt); err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	return nil
}

func notificationCopy(eventType, reference string) (string, string, bool) {
	switch eventType {
	case "purchase.created":
		return "Checkout started", reference + " inventory reserved while payment is pending.", true
	case "purchase.paid":
		return "Payment confirmed", reference + " is paid and ready for seller fulfillment.", true
	case "purchase.payment_failed":
		return "Payment failed", reference + " was not paid; reserved inventory was released.", true
	case "purchase.expired":
		return "Reservation expired", reference + " expired; reserved inventory was released.", true
	case "purchase.cancelled":
		return "Purchase cancelled", reference + " was cancelled and eligible inventory was restored.", true
	case "order.processing":
		return "Order processing", reference + " is being prepared by the seller.", true
	case "order.shipped":
		return "Order shipped", reference + " has left the seller.", true
	case "order.delivered":
		return "Order delivered", reference + " was delivered and can now be reviewed.", true
	case "order.cancelled":
		return "Order cancelled", reference + " seller order was cancelled and inventory restored.", true
	default:
		return "", "", false
	}
}
