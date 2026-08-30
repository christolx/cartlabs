package messaging

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNotificationCopy(t *testing.T) {
	tests := []struct {
		eventType string
		wantTitle string
		valid     bool
	}{
		{"purchase.created", "Checkout started", true},
		{"purchase.paid", "Payment confirmed", true},
		{"purchase.payment_failed", "Payment failed", true},
		{"purchase.expired", "Reservation expired", true},
		{"purchase.cancelled", "Purchase cancelled", true},
		{"order.processing", "Order processing", true},
		{"order.shipped", "Order shipped", true},
		{"order.delivered", "Order delivered", true},
		{"order.cancelled", "Order cancelled", true},
		{"product.updated", "", false},
	}
	for _, test := range tests {
		t.Run(test.eventType, func(t *testing.T) {
			title, body, valid := notificationCopy(test.eventType, "CL-01989F000000", "buyer")
			if valid != test.valid || title != test.wantTitle || valid && body == "" {
				t.Fatalf("title=%q body=%q valid=%v", title, body, valid)
			}
		})
	}
}

func TestNotificationCopyIsRoleAware(t *testing.T) {
	_, buyerBody, buyerValid := notificationCopy("order.delivered", "CL-01989F000000", "buyer")
	_, sellerBody, sellerValid := notificationCopy("order.delivered", "CL-01989F000000", "seller")
	if !buyerValid || !sellerValid || buyerBody == sellerBody {
		t.Fatalf("buyer=%q seller=%q", buyerBody, sellerBody)
	}
	if got := notificationHref("order.delivered", "purchase-1", "buyer"); got != "/purchases/purchase-1" {
		t.Fatalf("buyer href=%q", got)
	}
	if got := notificationHref("order.delivered", "purchase-1", "seller"); got != "/seller/orders" {
		t.Fatalf("seller href=%q", got)
	}
}

func TestRetryCount(t *testing.T) {
	for _, test := range []struct {
		value any
		want  int
	}{{int32(2), 2}, {int64(3), 3}, {int(4), 4}, {"invalid", 0}, {nil, 0}} {
		if got := retryCount(amqp.Table{"x-retry-count": test.value}); got != test.want {
			t.Fatalf("retryCount(%T(%v))=%d want=%d", test.value, test.value, got, test.want)
		}
	}
}
