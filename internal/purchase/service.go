package purchase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/google/uuid"
)

type Repository interface {
	Cart(context.Context, string) (Cart, error)
	SetCartItem(context.Context, string, string, int, time.Time) (Cart, error)
	RemoveCartItem(context.Context, string, string, time.Time) (Cart, error)
	Checkout(context.Context, string, string, time.Time, time.Time) (Purchase, bool, error)
	AttachPaymentIntent(context.Context, string, string, time.Time) (Purchase, error)
	FailPaymentSetup(context.Context, string, time.Time) error
	ListPurchases(context.Context, string) ([]Purchase, error)
	FindPurchase(context.Context, string, string) (Purchase, error)
	ListSellerOrders(context.Context, string) ([]SellerOrder, error)
	HandlePaymentEvent(context.Context, PaymentEvent, time.Time) (Purchase, error)
	ListNotifications(context.Context, string) ([]Notification, error)
	ExpireReservations(context.Context, time.Time) (int, error)
}

type PaymentProvider interface {
	CreateIntent(context.Context, PaymentIntentRequest) (PaymentIntent, error)
	CompleteIntent(context.Context, string, string) error
}

type Config struct {
	WebhookSecret  []byte
	WebhookURL     string
	ReservationTTL time.Duration
	DemoMode       bool
	Now            func() time.Time
}

type Service struct {
	repository     Repository
	provider       PaymentProvider
	webhookSecret  []byte
	webhookURL     string
	reservationTTL time.Duration
	demoMode       bool
	now            func() time.Time
}

func NewService(repository Repository, provider PaymentProvider, cfg Config) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("purchase repository is required")
	}
	if len(cfg.WebhookSecret) < 32 {
		return nil, fmt.Errorf("payment webhook secret must contain at least 32 bytes")
	}
	if strings.TrimSpace(cfg.WebhookURL) == "" {
		return nil, fmt.Errorf("payment webhook URL is required")
	}
	if cfg.ReservationTTL <= 0 {
		cfg.ReservationTTL = 15 * time.Minute
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Service{repository: repository, provider: provider, webhookSecret: cfg.WebhookSecret,
		webhookURL: cfg.WebhookURL, reservationTTL: cfg.ReservationTTL, demoMode: cfg.DemoMode, now: cfg.Now}, nil
}

func (s *Service) Cart(ctx context.Context, principal identity.Principal) (Cart, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Cart{}, domain.ErrForbidden
	}
	return s.repository.Cart(ctx, principal.UserID)
}

func (s *Service) SetCartItem(ctx context.Context, principal identity.Principal, variantID string, quantity int) (Cart, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Cart{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(variantID); err != nil || quantity < 1 || quantity > 99 {
		return Cart{}, domain.ErrInvalid
	}
	return s.repository.SetCartItem(ctx, principal.UserID, variantID, quantity, s.now().UTC())
}

func (s *Service) RemoveCartItem(ctx context.Context, principal identity.Principal, variantID string) (Cart, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Cart{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(variantID); err != nil {
		return Cart{}, domain.ErrInvalid
	}
	return s.repository.RemoveCartItem(ctx, principal.UserID, variantID, s.now().UTC())
}

func (s *Service) Checkout(ctx context.Context, principal identity.Principal, idempotencyKey string) (Purchase, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Purchase{}, domain.ErrForbidden
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) < 8 || len(idempotencyKey) > 128 || strings.ContainsAny(idempotencyKey, "\r\n") {
		return Purchase{}, domain.ErrInvalid
	}
	now := s.now().UTC()
	result, created, err := s.repository.Checkout(ctx, principal.UserID, idempotencyKey, now, now.Add(s.reservationTTL))
	if err != nil || !created {
		return result, err
	}
	if s.provider == nil {
		_ = s.repository.FailPaymentSetup(ctx, result.ID, now)
		return Purchase{}, domain.ErrUnavailable
	}
	intent, err := s.provider.CreateIntent(ctx, PaymentIntentRequest{Reference: result.Reference, AmountMinor: result.TotalMinor,
		Currency: result.Currency, WebhookURL: s.webhookURL})
	if err != nil {
		_ = s.repository.FailPaymentSetup(ctx, result.ID, now)
		return Purchase{}, domain.ErrUnavailable
	}
	if intent.Reference != result.Reference || intent.AmountMinor != result.TotalMinor || intent.Currency != result.Currency || intent.ID == "" {
		_ = s.repository.FailPaymentSetup(ctx, result.ID, now)
		return Purchase{}, domain.ErrUnavailable
	}
	return s.repository.AttachPaymentIntent(ctx, result.ID, intent.ID, now)
}

func (s *Service) ListPurchases(ctx context.Context, principal identity.Principal) ([]Purchase, error) {
	if !principal.Require(identity.RoleBuyer) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListPurchases(ctx, principal.UserID)
}

func (s *Service) Purchase(ctx context.Context, principal identity.Principal, purchaseID string) (Purchase, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Purchase{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(purchaseID); err != nil {
		return Purchase{}, domain.ErrInvalid
	}
	return s.repository.FindPurchase(ctx, principal.UserID, purchaseID)
}

func (s *Service) SellerOrders(ctx context.Context, principal identity.Principal) ([]SellerOrder, error) {
	if !principal.Require(identity.RoleSeller) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListSellerOrders(ctx, principal.UserID)
}

func (s *Service) ConfirmPayment(ctx context.Context, principal identity.Principal, purchaseID, outcome string) (Purchase, error) {
	if !principal.Require(identity.RoleBuyer) {
		return Purchase{}, domain.ErrForbidden
	}
	if !s.demoMode {
		return Purchase{}, domain.ErrForbidden
	}
	outcome = strings.ToLower(strings.TrimSpace(outcome))
	if _, err := uuid.Parse(purchaseID); err != nil || outcome != "succeeded" && outcome != "failed" {
		return Purchase{}, domain.ErrInvalid
	}
	current, err := s.repository.FindPurchase(ctx, principal.UserID, purchaseID)
	if err != nil {
		return Purchase{}, err
	}
	if current.Status != "pending_payment" || current.PaymentIntentID == "" {
		return Purchase{}, domain.ErrConflict
	}
	if s.provider == nil {
		return Purchase{}, domain.ErrUnavailable
	}
	if err := s.provider.CompleteIntent(ctx, current.PaymentIntentID, outcome); err != nil {
		return Purchase{}, domain.ErrUnavailable
	}
	return s.repository.FindPurchase(ctx, principal.UserID, purchaseID)
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	now := s.now().UTC()
	if err := VerifyWebhookSignature(s.webhookSecret, payload, signature, now, 5*time.Minute); err != nil {
		return domain.ErrUnauthorized
	}
	var event PaymentEvent
	decoderErr := json.Unmarshal(payload, &event)
	event.Payload = append([]byte(nil), payload...)
	if decoderErr != nil || event.ID == "" || event.Data.IntentID == "" || event.Data.Reference == "" ||
		event.Data.AmountMinor < 0 || event.Data.Currency != "IDR" ||
		event.Type != "payment.succeeded" && event.Type != "payment.failed" || event.CreatedAt.IsZero() {
		return domain.ErrInvalid
	}
	_, err := s.repository.HandlePaymentEvent(ctx, event, now)
	return err
}

func (s *Service) Notifications(ctx context.Context, principal identity.Principal) ([]Notification, error) {
	if !principal.Require(identity.RoleBuyer, identity.RoleSeller) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListNotifications(ctx, principal.UserID)
}

func (s *Service) ExpireReservations(ctx context.Context) (int, error) {
	return s.repository.ExpireReservations(ctx, s.now().UTC())
}

func IsTerminal(status string) bool {
	return status == "paid" || status == "payment_failed" || status == "expired"
}
