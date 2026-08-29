package httpapi

import (
	"fmt"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/purchase"
)

func toDomainReviewInput(value contract.ReviewInput) purchase.ReviewInput {
	return purchase.ReviewInput{PurchaseItemID: value.PurchaseItemId.String(), Rating: value.Rating, Title: value.Title, Body: value.Body}
}

func toContractCartItem(value purchase.CartItem) (contract.CartItem, error) {
	variantID, err := contractUUID(value.VariantID, "cartItem.variantId")
	if err != nil {
		return contract.CartItem{}, err
	}
	productID, err := contractUUID(value.ProductID, "cartItem.productId")
	if err != nil {
		return contract.CartItem{}, err
	}
	currency := contract.CartItemCurrency(value.Currency)
	if !currency.Valid() {
		return contract.CartItem{}, fmt.Errorf("map invalid cart item currency %q", value.Currency)
	}
	return contract.CartItem{
		VariantId: variantID, ProductId: productID, ProductName: value.ProductName,
		ProductSlug: value.ProductSlug, VariantName: value.VariantName, Sku: value.SKU,
		ImageUrl: value.ImageURL, Quantity: value.Quantity, AvailableStock: value.AvailableStock,
		UnitPriceMinor: value.UnitPriceMinor, LineTotalMinor: value.LineTotalMinor, Currency: currency,
	}, nil
}

func toContractCartStore(value purchase.CartStore) (contract.CartStore, error) {
	storeID, err := contractUUID(value.StoreID, "cartStore.storeId")
	if err != nil {
		return contract.CartStore{}, err
	}
	items := make([]contract.CartItem, len(value.Items))
	for i := range value.Items {
		mapped, err := toContractCartItem(value.Items[i])
		if err != nil {
			return contract.CartStore{}, err
		}
		items[i] = mapped
	}
	return contract.CartStore{StoreId: storeID, StoreName: value.StoreName, Items: items, SubtotalMinor: value.SubtotalMinor}, nil
}

func toContractCart(value purchase.Cart) (contract.Cart, error) {
	id, err := contractUUID(value.ID, "cart.id")
	if err != nil {
		return contract.Cart{}, err
	}
	currency := contract.CartCurrency(value.Currency)
	if !currency.Valid() {
		return contract.Cart{}, fmt.Errorf("map invalid cart currency %q", value.Currency)
	}
	stores := make([]contract.CartStore, len(value.Stores))
	for i := range value.Stores {
		mapped, err := toContractCartStore(value.Stores[i])
		if err != nil {
			return contract.Cart{}, err
		}
		stores[i] = mapped
	}
	return contract.Cart{Id: id, Currency: currency, Stores: stores, SubtotalMinor: value.SubtotalMinor, TotalQuantity: value.TotalQuantity, UpdatedAt: value.UpdatedAt}, nil
}

func toContractPurchaseItem(value purchase.PurchaseItem) (contract.PurchaseItem, error) {
	id, err := contractUUID(value.ID, "purchaseItem.id")
	if err != nil {
		return contract.PurchaseItem{}, err
	}
	productID, err := contractUUID(value.ProductID, "purchaseItem.productId")
	if err != nil {
		return contract.PurchaseItem{}, err
	}
	variantID, err := contractUUID(value.VariantID, "purchaseItem.variantId")
	if err != nil {
		return contract.PurchaseItem{}, err
	}
	currency := contract.PurchaseItemCurrency(value.Currency)
	if !currency.Valid() {
		return contract.PurchaseItem{}, fmt.Errorf("map invalid purchase item currency %q", value.Currency)
	}
	return contract.PurchaseItem{
		Id: id, ProductId: productID, VariantId: variantID, ProductName: value.ProductName,
		VariantName: value.VariantName, Sku: value.SKU, ImageUrl: value.ImageURL,
		Quantity: value.Quantity, UnitPriceMinor: value.UnitPriceMinor,
		LineTotalMinor: value.LineTotalMinor, Currency: currency, Reviewed: value.Reviewed,
	}, nil
}

func toContractSellerOrder(value purchase.SellerOrder) (contract.SellerOrder, error) {
	id, err := contractUUID(value.ID, "sellerOrder.id")
	if err != nil {
		return contract.SellerOrder{}, err
	}
	purchaseID, err := contractUUID(value.PurchaseID, "sellerOrder.purchaseId")
	if err != nil {
		return contract.SellerOrder{}, err
	}
	storeID, err := contractUUID(value.StoreID, "sellerOrder.storeId")
	if err != nil {
		return contract.SellerOrder{}, err
	}
	status := contract.SellerOrderStatus(value.Status)
	if !status.Valid() {
		return contract.SellerOrder{}, fmt.Errorf("map invalid seller order status %q", value.Status)
	}
	currency := contract.SellerOrderCurrency(value.Currency)
	if !currency.Valid() {
		return contract.SellerOrder{}, fmt.Errorf("map invalid seller order currency %q", value.Currency)
	}
	items := make([]contract.PurchaseItem, len(value.Items))
	for i := range value.Items {
		mapped, err := toContractPurchaseItem(value.Items[i])
		if err != nil {
			return contract.SellerOrder{}, err
		}
		items[i] = mapped
	}
	return contract.SellerOrder{
		Id: id, PurchaseId: purchaseID, Reference: value.Reference, StoreId: storeID,
		StoreName: value.StoreName, Status: status, Currency: currency,
		SubtotalMinor: value.SubtotalMinor, Items: items, ProcessingAt: value.ProcessingAt,
		ShippedAt: value.ShippedAt, DeliveredAt: value.DeliveredAt, CancelledAt: value.CancelledAt,
		CancellationReason: value.CancellationReason, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}, nil
}

func toContractSellerOrders(values []purchase.SellerOrder) ([]contract.SellerOrder, error) {
	result := make([]contract.SellerOrder, len(values))
	for i := range values {
		mapped, err := toContractSellerOrder(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractPurchase(value purchase.Purchase) (contract.Purchase, error) {
	id, err := contractUUID(value.ID, "purchase.id")
	if err != nil {
		return contract.Purchase{}, err
	}
	buyerID, err := contractUUID(value.BuyerID, "purchase.buyerId")
	if err != nil {
		return contract.Purchase{}, err
	}
	status := contract.PurchaseStatus(value.Status)
	if !status.Valid() {
		return contract.Purchase{}, fmt.Errorf("map invalid purchase status %q", value.Status)
	}
	paymentStatus := contract.PaymentStatus(value.PaymentStatus)
	if !paymentStatus.Valid() {
		return contract.Purchase{}, fmt.Errorf("map invalid payment status %q", value.PaymentStatus)
	}
	currency := contract.PurchaseCurrency(value.Currency)
	if !currency.Valid() {
		return contract.Purchase{}, fmt.Errorf("map invalid purchase currency %q", value.Currency)
	}
	orders, err := toContractSellerOrders(value.SellerOrders)
	if err != nil {
		return contract.Purchase{}, err
	}
	return contract.Purchase{
		Id: id, Reference: value.Reference, BuyerId: buyerID, Status: status,
		PaymentStatus: paymentStatus, PaymentIntentId: value.PaymentIntentID, Currency: currency,
		SubtotalMinor: value.SubtotalMinor, TotalMinor: value.TotalMinor,
		ReservationExpiresAt: value.ReservationExpiresAt, SellerOrders: orders,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}, nil
}

func toContractPurchases(values []purchase.Purchase) ([]contract.Purchase, error) {
	result := make([]contract.Purchase, len(values))
	for i := range values {
		mapped, err := toContractPurchase(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractReview(value purchase.Review) (contract.Review, error) {
	id, err := contractUUID(value.ID, "review.id")
	if err != nil {
		return contract.Review{}, err
	}
	buyerID, err := contractUUID(value.BuyerID, "review.buyerId")
	if err != nil {
		return contract.Review{}, err
	}
	itemID, err := contractUUID(value.PurchaseItemID, "review.purchaseItemId")
	if err != nil {
		return contract.Review{}, err
	}
	productID, err := contractUUID(value.ProductID, "review.productId")
	if err != nil {
		return contract.Review{}, err
	}
	return contract.Review{Id: id, BuyerId: buyerID, BuyerName: value.BuyerName, PurchaseItemId: itemID, ProductId: productID, Rating: value.Rating, Title: value.Title, Body: value.Body, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}, nil
}

func toContractReviews(values []purchase.Review) ([]contract.Review, error) {
	result := make([]contract.Review, len(values))
	for i := range values {
		mapped, err := toContractReview(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractReviewSummary(value purchase.ReviewSummary) (contract.ReviewSummary, error) {
	items, err := toContractReviews(value.Items)
	if err != nil {
		return contract.ReviewSummary{}, err
	}
	return contract.ReviewSummary{Average: value.Average, Count: value.Count, Items: items}, nil
}

func toContractAdminOverview(value purchase.AdminOverview) (contract.AdminOverview, error) {
	currency := contract.AdminOverviewCurrency(value.Currency)
	if !currency.Valid() {
		return contract.AdminOverview{}, fmt.Errorf("map invalid admin overview currency %q", value.Currency)
	}
	return contract.AdminOverview{
		Users: value.Users, ApprovedStores: value.ApprovedStores, PublishedProducts: value.PublishedProducts,
		Purchases: value.Purchases, ActiveSellerOrders: value.ActiveSellerOrders,
		DeliveredOrders: value.DeliveredOrders, GrossMerchandiseMinor: value.GrossMerchandiseMinor,
		Currency: currency,
	}, nil
}

func toContractAuditEvent(value purchase.AuditEvent) (contract.AuditEvent, error) {
	id, err := contractUUID(value.ID, "auditEvent.id")
	if err != nil {
		return contract.AuditEvent{}, err
	}
	resourceID, err := contractUUID(value.ResourceID, "auditEvent.resourceId")
	if err != nil {
		return contract.AuditEvent{}, err
	}
	role := contract.AuditEventActorRole(value.ActorRole)
	if !role.Valid() {
		return contract.AuditEvent{}, fmt.Errorf("map invalid audit actor role %q", value.ActorRole)
	}
	data := value.Data
	if data == nil {
		data = map[string]any{}
	}
	result := contract.AuditEvent{Id: id, ActorRole: role, Action: value.Action, ResourceType: value.ResourceType, ResourceId: resourceID, Data: data, CreatedAt: value.CreatedAt}
	if value.ActorID != "" {
		actorID, err := contractUUID(value.ActorID, "auditEvent.actorId")
		if err != nil {
			return contract.AuditEvent{}, err
		}
		result.ActorId = &actorID
	}
	if value.ActorName != "" {
		result.ActorName = &value.ActorName
	}
	return result, nil
}

func toContractAuditEvents(values []purchase.AuditEvent) ([]contract.AuditEvent, error) {
	result := make([]contract.AuditEvent, len(values))
	for i := range values {
		mapped, err := toContractAuditEvent(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractNotification(value purchase.Notification) (contract.Notification, error) {
	id, err := contractUUID(value.ID, "notification.id")
	if err != nil {
		return contract.Notification{}, err
	}
	return contract.Notification{Id: id, Kind: value.Kind, Title: value.Title, Body: value.Body, ReadAt: value.ReadAt, CreatedAt: value.CreatedAt}, nil
}

func toContractNotifications(values []purchase.Notification) ([]contract.Notification, error) {
	result := make([]contract.Notification, len(values))
	for i := range values {
		mapped, err := toContractNotification(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}
