package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	DefaultExchange  = "cartlabs.events"
	DefaultQueue     = "cartlabs.notifications.v2"
	DeadExchange     = "cartlabs.dead"
	DeadQueue        = "cartlabs.notifications.dead"
	RetryExchange    = "cartlabs.retry"
	RetryQueue       = "cartlabs.notifications.retry"
	SearchQueue      = "cartlabs.search.v1"
	SearchDeadQueue  = "cartlabs.search.dead"
	SearchRetryQueue = "cartlabs.search.retry"
	maxAttempts      = 5
	maxDeliveries    = 3
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
	ctx, span := otel.Tracer("github.com/christolx/cartlabs/messaging").Start(ctx, "outbox.relay")
	defer span.End()
	tx, err := o.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin outbox relay: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var event Event
	var payload []byte
	var attempts int
	err = tx.QueryRow(ctx, `
		SELECT id::text,event_type,occurred_at,payload,attempts FROM outbox_events
		WHERE published_at IS NULL AND dead_lettered_at IS NULL AND next_attempt_at <= $1
		ORDER BY next_attempt_at,occurred_at,id FOR UPDATE SKIP LOCKED LIMIT 1`, now).
		Scan(&event.ID, &event.Type, &event.OccurredAt, &payload, &attempts)
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
		attempts++
		delay := time.Second * time.Duration(1<<min(attempts, 6))
		if _, updateErr := tx.Exec(ctx, `UPDATE outbox_events SET attempts=$2::integer,last_error=$3,next_attempt_at=$4,
			dead_lettered_at=CASE WHEN $2::integer >= $5::integer THEN $6::timestamptz ELSE dead_lettered_at END WHERE id=$1`,
			event.ID, attempts, message, now.Add(delay), maxAttempts, now); updateErr != nil {
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

func (o *Outbox) ReplayDeadLetters(ctx context.Context, limit int, now time.Time) (int64, error) {
	if limit < 1 || limit > 1000 {
		return 0, domain.ErrInvalid
	}
	result, err := o.pool.Exec(ctx, `WITH selected AS (
		SELECT id FROM outbox_events WHERE dead_lettered_at IS NOT NULL ORDER BY dead_lettered_at,id LIMIT $1 FOR UPDATE SKIP LOCKED
	) UPDATE outbox_events o SET attempts=0,last_error='',next_attempt_at=$2,dead_lettered_at=NULL
	FROM selected WHERE o.id=selected.id`, limit, now)
	if err != nil {
		return 0, fmt.Errorf("replay outbox dead letters: %w", err)
	}
	return result.RowsAffected(), nil
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
	ctx, span := otel.Tracer("github.com/christolx/cartlabs/messaging").Start(ctx, "rabbitmq.publish "+routingKey,
		trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(attribute.String("messaging.system", "rabbitmq"), attribute.String("messaging.destination.name", p.exchange)))
	defer span.End()
	if err := p.channel.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json", DeliveryMode: amqp.Persistent, Timestamp: time.Now().UTC(), Body: body, Headers: injectTrace(ctx, nil),
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	channel       *amqp.Channel
	messages      <-chan amqp.Delivery
	confirmations <-chan amqp.Confirmation
	retryExchange string
	retryKey      string
	name          string
}

type ConsumerConfig struct {
	Exchange      string
	Queue         string
	Consumer      string
	Bindings      []string
	RetryExchange string
	RetryQueue    string
	RetryKey      string
	DeadExchange  string
	DeadQueue     string
	DeadKey       string
}

func NewNotificationConsumer(connection *amqp.Connection, exchange, queue string) (*Consumer, error) {
	return NewConsumer(connection, ConsumerConfig{Exchange: exchange, Queue: queue, Consumer: "cartlabs-worker-notifications",
		Bindings: []string{"purchase.#", "order.#", "notifications.retry"}, RetryExchange: RetryExchange,
		RetryQueue: RetryQueue, RetryKey: "notifications.retry", DeadExchange: DeadExchange, DeadQueue: DeadQueue, DeadKey: "notifications.failed"})
}

func NewSearchConsumer(connection *amqp.Connection, exchange string) (*Consumer, error) {
	return NewConsumer(connection, ConsumerConfig{Exchange: exchange, Queue: SearchQueue, Consumer: "cartlabs-worker-search",
		Bindings: []string{"catalog.search.#", "search.retry"}, RetryExchange: RetryExchange,
		RetryQueue: SearchRetryQueue, RetryKey: "search.retry", DeadExchange: DeadExchange, DeadQueue: SearchDeadQueue, DeadKey: "search.failed"})
}

func NewConsumer(connection *amqp.Connection, cfg ConsumerConfig) (*Consumer, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open consumer channel: %w", err)
	}
	closeWith := func(err error) (*Consumer, error) {
		_ = channel.Close()
		return nil, err
	}
	if err := channel.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare consumer exchange: %w", err))
	}
	if err := channel.ExchangeDeclare(cfg.DeadExchange, "direct", true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare dead-letter exchange: %w", err))
	}
	if err := channel.ExchangeDeclare(cfg.RetryExchange, "direct", true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare retry exchange: %w", err))
	}
	if _, err := channel.QueueDeclare(cfg.DeadQueue, true, false, false, false, nil); err != nil {
		return closeWith(fmt.Errorf("declare dead-letter queue: %w", err))
	}
	if err := channel.QueueBind(cfg.DeadQueue, cfg.DeadKey, cfg.DeadExchange, false, nil); err != nil {
		return closeWith(fmt.Errorf("bind dead-letter queue: %w", err))
	}
	if _, err := channel.QueueDeclare(cfg.RetryQueue, true, false, false, false, amqp.Table{
		"x-message-ttl": int32(2000), "x-dead-letter-exchange": cfg.Exchange,
	}); err != nil {
		return closeWith(fmt.Errorf("declare retry queue: %w", err))
	}
	if err := channel.QueueBind(cfg.RetryQueue, cfg.RetryKey, cfg.RetryExchange, false, nil); err != nil {
		return closeWith(fmt.Errorf("bind retry queue: %w", err))
	}
	if _, err := channel.QueueDeclare(cfg.Queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": cfg.DeadExchange, "x-dead-letter-routing-key": cfg.DeadKey,
	}); err != nil {
		return closeWith(fmt.Errorf("declare %s queue: %w", cfg.Consumer, err))
	}
	for _, routingKey := range cfg.Bindings {
		if err := channel.QueueBind(cfg.Queue, routingKey, cfg.Exchange, false, nil); err != nil {
			return closeWith(fmt.Errorf("bind %s queue: %w", cfg.Consumer, err))
		}
	}
	if err := channel.Qos(10, 0, false); err != nil {
		return closeWith(fmt.Errorf("configure consumer QoS: %w", err))
	}
	if err := channel.Confirm(false); err != nil {
		return closeWith(fmt.Errorf("enable retry publisher confirms: %w", err))
	}
	confirmations := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	messages, err := channel.Consume(cfg.Queue, cfg.Consumer, false, false, false, false, nil)
	if err != nil {
		return closeWith(fmt.Errorf("consume notification queue: %w", err))
	}
	return &Consumer{channel: channel, messages: messages, confirmations: confirmations,
		retryExchange: cfg.RetryExchange, retryKey: cfg.RetryKey, name: cfg.Consumer}, nil
}

type Handler interface {
	Handle(context.Context, []byte) error
}

func (c *Consumer) Run(ctx context.Context, handler Handler, logger *slog.Logger, observers ...func(string)) error {
	var observe func(string)
	if len(observers) > 0 {
		observe = observers[0]
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-c.messages:
			if !ok {
				return fmt.Errorf("%s delivery channel closed", c.name)
			}
			deliveryCtx := extractTrace(ctx, delivery.Headers)
			deliveryCtx, span := otel.Tracer("github.com/christolx/cartlabs/messaging").Start(deliveryCtx, "rabbitmq.consume "+delivery.RoutingKey,
				trace.WithSpanKind(trace.SpanKindConsumer), trace.WithAttributes(attribute.String("messaging.system", "rabbitmq"), attribute.String("messaging.destination.name", delivery.RoutingKey)))
			err := handler.Handle(deliveryCtx, delivery.Body)
			if err == nil {
				span.End()
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack %s event: %w", c.name, ackErr)
				}
				observeResult(observe, "success")
				continue
			}
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			permanent := errors.Is(err, domain.ErrInvalid)
			attempts := retryCount(delivery.Headers)
			if !permanent && attempts < maxDeliveries {
				if retryErr := c.retry(deliveryCtx, delivery, attempts+1); retryErr != nil {
					span.End()
					if nackErr := delivery.Nack(false, true); nackErr != nil {
						return fmt.Errorf("requeue %s after retry publish failure: %w", c.name, nackErr)
					}
					return fmt.Errorf("publish %s retry: %w", c.name, retryErr)
				}
				span.End()
				if ackErr := delivery.Ack(false); ackErr != nil {
					return fmt.Errorf("ack retried %s event: %w", c.name, ackErr)
				}
				observeResult(observe, "retry")
				continue
			}
			span.End()
			logger.ErrorContext(deliveryCtx, "consumer event dead-lettered", "consumer", c.name, "error", err, "attempts", attempts)
			if nackErr := delivery.Nack(false, false); nackErr != nil {
				return fmt.Errorf("nack %s event: %w", c.name, nackErr)
			}
			observeResult(observe, "dead_letter")
		}
	}
}

func (c *Consumer) retry(ctx context.Context, delivery amqp.Delivery, attempts int) error {
	headers := amqp.Table{}
	for key, value := range delivery.Headers {
		headers[key] = value
	}
	headers["x-retry-count"] = int32(attempts)
	headers = injectTrace(ctx, headers)
	if err := c.channel.PublishWithContext(ctx, c.retryExchange, c.retryKey, false, false, amqp.Publishing{
		ContentType: delivery.ContentType, DeliveryMode: amqp.Persistent, Timestamp: time.Now().UTC(), Headers: headers, Body: delivery.Body,
	}); err != nil {
		return err
	}
	select {
	case confirmation, ok := <-c.confirmations:
		if !ok || !confirmation.Ack {
			return fmt.Errorf("broker did not confirm retry")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func retryCount(headers amqp.Table) int {
	switch value := headers["x-retry-count"].(type) {
	case int32:
		return int(value)
	case int64:
		return int(value)
	case int:
		return value
	default:
		return 0
	}
}

func observeResult(observe func(string), result string) {
	if observe != nil {
		observe(result)
	}
}

func injectTrace(ctx context.Context, headers amqp.Table) amqp.Table {
	if headers == nil {
		headers = amqp.Table{}
	}
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for key, value := range carrier {
		headers[key] = value
	}
	return headers
}

func extractTrace(ctx context.Context, headers amqp.Table) context.Context {
	carrier := propagation.MapCarrier{}
	for key, value := range headers {
		if text, ok := value.(string); ok {
			carrier[key] = text
		}
	}
	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}

type DeadLetterReplayer struct {
	channel       *amqp.Channel
	confirmations <-chan amqp.Confirmation
	exchange      string
}

func NewDeadLetterReplayer(connection *amqp.Connection, exchange string) (*DeadLetterReplayer, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("open dead-letter replay channel: %w", err)
	}
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("enable dead-letter replay confirms: %w", err)
	}
	if err := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare replay exchange: %w", err)
	}
	if err := channel.ExchangeDeclare(DeadExchange, "direct", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare replay dead-letter exchange: %w", err)
	}
	if _, err := channel.QueueDeclare(DeadQueue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("declare replay dead-letter queue: %w", err)
	}
	if err := channel.QueueBind(DeadQueue, "notifications.failed", DeadExchange, false, nil); err != nil {
		_ = channel.Close()
		return nil, fmt.Errorf("bind replay dead-letter queue: %w", err)
	}
	return &DeadLetterReplayer{channel: channel, confirmations: channel.NotifyPublish(make(chan amqp.Confirmation, 1)), exchange: exchange}, nil
}

func (r *DeadLetterReplayer) Replay(ctx context.Context, limit int) (int, error) {
	if limit < 1 || limit > 1000 {
		return 0, domain.ErrInvalid
	}
	replayed := 0
	for replayed < limit {
		delivery, ok, err := r.channel.Get(DeadQueue, false)
		if err != nil {
			return replayed, fmt.Errorf("get dead-letter notification: %w", err)
		}
		if !ok {
			break
		}
		var event Event
		if err := json.Unmarshal(delivery.Body, &event); err != nil || event.Type == "" {
			if rejectErr := delivery.Reject(false); rejectErr != nil {
				return replayed, fmt.Errorf("reject invalid dead letter: %w", rejectErr)
			}
			continue
		}
		headers := delivery.Headers
		delete(headers, "x-retry-count")
		if err := r.channel.PublishWithContext(ctx, r.exchange, event.Type, false, false, amqp.Publishing{
			ContentType: "application/json", DeliveryMode: amqp.Persistent, Timestamp: time.Now().UTC(), Headers: injectTrace(ctx, headers), Body: delivery.Body,
		}); err != nil {
			_ = delivery.Nack(false, true)
			return replayed, fmt.Errorf("republish dead-letter notification: %w", err)
		}
		select {
		case confirmation, open := <-r.confirmations:
			if !open || !confirmation.Ack {
				_ = delivery.Nack(false, true)
				return replayed, fmt.Errorf("broker did not confirm dead-letter replay")
			}
		case <-ctx.Done():
			_ = delivery.Nack(false, true)
			return replayed, ctx.Err()
		}
		if err := delivery.Ack(false); err != nil {
			return replayed, fmt.Errorf("ack replayed dead letter: %w", err)
		}
		replayed++
	}
	return replayed, nil
}

func (r *DeadLetterReplayer) Close() error { return r.channel.Close() }

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
	for _, recipientID := range data.RecipientIDs {
		var role string
		if err := h.pool.QueryRow(ctx, `SELECT role::text FROM users WHERE id=$1`, recipientID).Scan(&role); errors.Is(err, pgx.ErrNoRows) {
			continue
		} else if err != nil {
			return fmt.Errorf("find notification recipient: %w", err)
		}
		title, message, ok := notificationCopy(event.Type, data.Reference, role)
		if !ok {
			return domain.ErrInvalid
		}
		href := notificationHref(event.Type, data.PurchaseID, role)
		notificationID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("generate notification ID: %w", err)
		}
		if _, err := h.pool.Exec(ctx, `
			INSERT INTO notifications (id,user_id,source_event_id,kind,title,body,href,created_at)
			SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE EXISTS (SELECT 1 FROM users WHERE id=$2)
			ON CONFLICT (user_id,source_event_id) DO NOTHING`, notificationID.String(), recipientID, event.ID,
			event.Type, title, message, href, event.OccurredAt); err != nil {
			return fmt.Errorf("insert notification: %w", err)
		}
	}
	return nil
}

func notificationCopy(eventType, reference, role string) (string, string, bool) {
	switch eventType {
	case "purchase.created":
		return "Checkout started", reference + " inventory reserved while payment is pending.", true
	case "purchase.paid":
		if role == "seller" {
			return "Payment confirmed", reference + " is paid. Fulfillment can begin.", true
		}
		return "Payment confirmed", reference + " is paid and ready for seller fulfillment.", true
	case "purchase.payment_failed":
		return "Payment failed", reference + " was not paid; reserved inventory was released.", true
	case "purchase.expired":
		return "Reservation expired", reference + " expired; reserved inventory was released.", true
	case "purchase.cancelled":
		return "Purchase cancelled", reference + " was cancelled and eligible inventory was restored.", true
	case "order.processing":
		if role == "seller" {
			return "Order processing", reference + " is now being prepared.", true
		}
		return "Order processing", reference + " is being prepared by the seller.", true
	case "order.shipped":
		if role == "seller" {
			return "Order shipped", reference + " was marked shipped.", true
		}
		return "Order shipped", reference + " has left the seller.", true
	case "order.delivered":
		if role == "seller" {
			return "Order delivered", reference + " was delivered. No seller action is required.", true
		}
		return "Order delivered", reference + " was delivered and can now be reviewed.", true
	case "order.cancelled":
		if role == "seller" {
			return "Order cancelled", reference + " was cancelled. Reserved inventory was restored.", true
		}
		return "Order cancelled", reference + " seller order was cancelled and inventory restored.", true
	default:
		return "", "", false
	}
}

func notificationHref(eventType, purchaseID, role string) string {
	if !strings.HasPrefix(eventType, "purchase.") && !strings.HasPrefix(eventType, "order.") {
		return ""
	}
	if role == "seller" {
		return "/seller/orders"
	}
	if role == "buyer" {
		return "/purchases/" + purchaseID
	}
	return ""
}
