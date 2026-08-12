package httpapi

import (
	"io"
	"net/http"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
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
	response, err := toContractCart(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) setCartItem(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input contract.CartItemInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.SetCartItem(r.Context(), principal, r.PathValue("variantId"), input.Quantity)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractCart(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractCart(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractPurchase(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
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
	response, err := toContractPurchases(items)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.Purchase]{Items: response})
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
	response, err := toContractPurchase(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractSellerOrders(items)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.SellerOrder]{Items: response})
}

func (a *api) confirmPayment(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input contract.PaymentCompletionInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.ConfirmPayment(r.Context(), principal, r.PathValue("purchaseId"), string(input.Outcome))
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractPurchase(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractNotifications(items)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.Notification]{Items: response})
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
	var input contract.FulfillmentInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	reason := ""
	if input.Reason != nil {
		reason = *input.Reason
	}
	result, err := a.config.purchases.UpdateSellerOrder(r.Context(), principal, r.PathValue("orderId"), string(input.Status), reason)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractSellerOrder(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) cancelPurchase(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input contract.CancellationInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.CancelPurchase(r.Context(), principal, r.PathValue("purchaseId"), input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractPurchase(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) createReview(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.purchases == nil {
		writeError(w, domain.ErrUnavailable)
		return
	}
	var input contract.ReviewInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	result, err := a.config.purchases.CreateReview(r.Context(), principal, toDomainReviewInput(input))
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractReview(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
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
	response, err := toContractReviewSummary(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractAdminOverview(result)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractAuditEvents(items)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.AuditEvent]{Items: response})
}
