package purchase

import "time"

type Cart struct {
	ID            string      `json:"id"`
	Currency      string      `json:"currency"`
	Stores        []CartStore `json:"stores"`
	SubtotalMinor int64       `json:"subtotalMinor"`
	TotalQuantity int         `json:"totalQuantity"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

type CartStore struct {
	StoreID       string     `json:"storeId"`
	StoreName     string     `json:"storeName"`
	Items         []CartItem `json:"items"`
	SubtotalMinor int64      `json:"subtotalMinor"`
}

type CartItem struct {
	VariantID      string `json:"variantId"`
	ProductID      string `json:"productId"`
	ProductName    string `json:"productName"`
	ProductSlug    string `json:"productSlug"`
	VariantName    string `json:"variantName"`
	SKU            string `json:"sku"`
	ImageURL       string `json:"imageUrl"`
	Quantity       int    `json:"quantity"`
	AvailableStock int    `json:"availableStock"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
	LineTotalMinor int64  `json:"lineTotalMinor"`
	Currency       string `json:"currency"`
}

type Purchase struct {
	ID                   string        `json:"id"`
	Reference            string        `json:"reference"`
	BuyerID              string        `json:"buyerId"`
	Status               string        `json:"status"`
	PaymentStatus        string        `json:"paymentStatus"`
	PaymentIntentID      string        `json:"paymentIntentId"`
	Currency             string        `json:"currency"`
	SubtotalMinor        int64         `json:"subtotalMinor"`
	TotalMinor           int64         `json:"totalMinor"`
	ReservationExpiresAt time.Time     `json:"reservationExpiresAt"`
	SellerOrders         []SellerOrder `json:"sellerOrders"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
}

type SellerOrder struct {
	ID            string         `json:"id"`
	PurchaseID    string         `json:"purchaseId"`
	Reference     string         `json:"reference"`
	StoreID       string         `json:"storeId"`
	StoreName     string         `json:"storeName"`
	Status        string         `json:"status"`
	Currency      string         `json:"currency"`
	SubtotalMinor int64          `json:"subtotalMinor"`
	Items         []PurchaseItem `json:"items"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type PurchaseItem struct {
	ID             string `json:"id"`
	ProductID      string `json:"productId"`
	VariantID      string `json:"variantId"`
	ProductName    string `json:"productName"`
	VariantName    string `json:"variantName"`
	SKU            string `json:"sku"`
	ImageURL       string `json:"imageUrl"`
	Quantity       int    `json:"quantity"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
	LineTotalMinor int64  `json:"lineTotalMinor"`
	Currency       string `json:"currency"`
}

type Notification struct {
	ID        string     `json:"id"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type PaymentIntentRequest struct {
	Reference   string `json:"reference"`
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
	WebhookURL  string `json:"webhookUrl"`
}

type PaymentIntent struct {
	ID          string `json:"id"`
	Reference   string `json:"reference"`
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
	Status      string `json:"status"`
}

type PaymentEvent struct {
	ID        string           `json:"id"`
	Type      string           `json:"type"`
	CreatedAt time.Time        `json:"createdAt"`
	Data      PaymentEventData `json:"data"`
	Payload   []byte           `json:"-"`
}

type PaymentEventData struct {
	IntentID    string `json:"intentId"`
	Reference   string `json:"reference"`
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}
