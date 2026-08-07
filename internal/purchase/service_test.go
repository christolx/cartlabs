package purchase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

type fakeRepository struct {
	checkoutResult  Purchase
	checkoutCreated bool
	checkoutErr     error
	attachedIntent  string
	failedSetup     bool
	handledEvent    PaymentEvent
	findResult      Purchase
}

func (f *fakeRepository) Cart(context.Context, string) (Cart, error) { return Cart{}, nil }
func (f *fakeRepository) SetCartItem(context.Context, string, string, int, time.Time) (Cart, error) {
	return Cart{}, nil
}
func (f *fakeRepository) RemoveCartItem(context.Context, string, string, time.Time) (Cart, error) {
	return Cart{}, nil
}
func (f *fakeRepository) Checkout(context.Context, string, string, time.Time, time.Time) (Purchase, bool, error) {
	return f.checkoutResult, f.checkoutCreated, f.checkoutErr
}
func (f *fakeRepository) AttachPaymentIntent(_ context.Context, _ string, intentID string, _ time.Time) (Purchase, error) {
	f.attachedIntent = intentID
	result := f.checkoutResult
	result.PaymentIntentID = intentID
	return result, nil
}
func (f *fakeRepository) FailPaymentSetup(context.Context, string, time.Time) error {
	f.failedSetup = true
	return nil
}
func (f *fakeRepository) ListPurchases(context.Context, string) ([]Purchase, error) { return nil, nil }
func (f *fakeRepository) FindPurchase(context.Context, string, string) (Purchase, error) {
	return f.findResult, nil
}
func (f *fakeRepository) ListSellerOrders(context.Context, string) ([]SellerOrder, error) {
	return nil, nil
}
func (f *fakeRepository) HandlePaymentEvent(_ context.Context, event PaymentEvent, _ time.Time) (Purchase, error) {
	f.handledEvent = event
	return Purchase{}, nil
}
func (f *fakeRepository) ListNotifications(context.Context, string) ([]Notification, error) {
	return nil, nil
}
func (f *fakeRepository) ExpireReservations(context.Context, time.Time) (int, error) { return 0, nil }

type fakeProvider struct {
	createResult PaymentIntent
	createErr    error
	createCalls  int
	completedID  string
	outcome      string
}

func (f *fakeProvider) CreateIntent(_ context.Context, _ PaymentIntentRequest) (PaymentIntent, error) {
	f.createCalls++
	return f.createResult, f.createErr
}
func (f *fakeProvider) CompleteIntent(_ context.Context, id, outcome string) error {
	f.completedID, f.outcome = id, outcome
	return nil
}

func newTestService(t *testing.T, repository Repository, provider PaymentProvider, now time.Time) *Service {
	t.Helper()
	service, err := NewService(repository, provider, Config{WebhookSecret: []byte("01234567890123456789012345678901"),
		WebhookURL: "http://api.test/payments/webhook", ReservationTTL: 15 * time.Minute, DemoMode: true,
		Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestCheckoutCreatesPaymentIntent(t *testing.T) {
	now := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	purchaseResult := Purchase{ID: "01989f00-0000-7000-8000-000000000701", Reference: "CL-01989F000000",
		Currency: "IDR", TotalMinor: 408000}
	repository := &fakeRepository{checkoutResult: purchaseResult, checkoutCreated: true}
	provider := &fakeProvider{createResult: PaymentIntent{ID: "intent-1", Reference: purchaseResult.Reference,
		AmountMinor: purchaseResult.TotalMinor, Currency: "IDR", Status: "pending"}}
	service := newTestService(t, repository, provider, now)

	result, err := service.Checkout(context.Background(), identity.Principal{UserID: "buyer", Role: identity.RoleBuyer}, "checkout-key-1")
	if err != nil {
		t.Fatal(err)
	}
	if provider.createCalls != 1 || repository.attachedIntent != "intent-1" || result.PaymentIntentID != "intent-1" {
		t.Fatalf("calls=%d attached=%q result=%#v", provider.createCalls, repository.attachedIntent, result)
	}
}

func TestCheckoutIdempotentResultSkipsProvider(t *testing.T) {
	repository := &fakeRepository{checkoutResult: Purchase{ID: "existing"}, checkoutCreated: false}
	provider := &fakeProvider{}
	service := newTestService(t, repository, provider, time.Now())
	result, err := service.Checkout(context.Background(), identity.Principal{UserID: "buyer", Role: identity.RoleBuyer}, "checkout-key-1")
	if err != nil || result.ID != "existing" || provider.createCalls != 0 {
		t.Fatalf("result=%#v err=%v calls=%d", result, err, provider.createCalls)
	}
}

func TestCheckoutProviderFailureReleasesReservation(t *testing.T) {
	repository := &fakeRepository{checkoutResult: Purchase{ID: "purchase", Reference: "CL-01989F000000", Currency: "IDR", TotalMinor: 1}, checkoutCreated: true}
	provider := &fakeProvider{createErr: errors.New("offline")}
	service := newTestService(t, repository, provider, time.Now())
	_, err := service.Checkout(context.Background(), identity.Principal{UserID: "buyer", Role: identity.RoleBuyer}, "checkout-key-1")
	if !errors.Is(err, domain.ErrUnavailable) || !repository.failedSetup {
		t.Fatalf("err=%v failedSetup=%v", err, repository.failedSetup)
	}
}

func TestCheckoutRequiresBuyer(t *testing.T) {
	service := newTestService(t, &fakeRepository{}, &fakeProvider{}, time.Now())
	_, err := service.Checkout(context.Background(), identity.Principal{Role: identity.RoleSeller}, "checkout-key-1")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err=%v", err)
	}
}

func TestHandleWebhookVerifiesSignature(t *testing.T) {
	now := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := newTestService(t, repository, nil, now)
	payload := []byte(`{"id":"event-1","type":"payment.succeeded","createdAt":"2026-08-08T00:00:00Z","data":{"intentId":"intent-1","reference":"CL-01989F000000","amountMinor":408000,"currency":"IDR"}}`)
	signature := SignWebhook([]byte("01234567890123456789012345678901"), payload, now)
	if err := service.HandleWebhook(context.Background(), payload, signature); err != nil {
		t.Fatal(err)
	}
	if repository.handledEvent.ID != "event-1" || string(repository.handledEvent.Payload) != string(payload) {
		t.Fatalf("event=%#v", repository.handledEvent)
	}
	if err := service.HandleWebhook(context.Background(), payload, signature+"00"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("invalid signature err=%v", err)
	}
}

func TestVerifyWebhookRejectsExpiredTimestamp(t *testing.T) {
	secret := []byte("01234567890123456789012345678901")
	payload := []byte(`{"id":"event"}`)
	signedAt := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	signature := SignWebhook(secret, payload, signedAt)
	if err := VerifyWebhookSignature(secret, payload, signature, signedAt.Add(6*time.Minute), 5*time.Minute); !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("err=%v", err)
	}
}
