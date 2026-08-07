package httpapi

import (
	"io"
	"net/http"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/christolx/cartlabs/internal/purchase"
)

func (a *api) getCart(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.Cart(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) setCartItem(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input struct {
		Quantity int `json:"quantity"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.SetCartItem(r.Context(), principal, r.PathValue("variantId"), input.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) removeCartItem(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.RemoveCartItem(r.Context(), principal, r.PathValue("variantId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) checkout(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.Checkout(r.Context(), principal, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *api) listPurchases(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	items, err := a.config.purchases.ListPurchases(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) getPurchase(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.Purchase(r.Context(), principal, r.PathValue("purchaseId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) sellerOrders(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	items, err := a.config.purchases.SellerOrders(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) confirmPayment(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input struct {
		Outcome string `json:"outcome"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.ConfirmPayment(r.Context(), principal, r.PathValue("purchaseId"), input.Outcome)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) notifications(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	items, err := a.config.purchases.Notifications(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) paymentWebhook(w http.ResponseWriter, r *http.Request) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		writeError(w, domain.ErrInvalid)
		return
	}
	if err := a.config.purchases.HandleWebhook(r.Context(), payload, r.Header.Get("X-Cartlabs-Signature")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) updateSellerOrder(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.UpdateSellerOrder(r.Context(), principal, r.PathValue("orderId"), input.Status, input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) cancelPurchase(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.CancelPurchase(r.Context(), principal, r.PathValue("purchaseId"), input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) createReview(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input purchase.ReviewInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.CreateReview(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (a *api) productReviews(w http.ResponseWriter, r *http.Request) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.Reviews(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) adminOverview(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	result, err := a.config.purchases.AdminOverview(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *api) auditEvents(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	items, err := a.config.purchases.AuditEvents(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
