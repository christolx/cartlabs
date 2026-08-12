package httpapi

import (
	"net/http"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

func (a *api) getStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	if a.config.stores == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	value, err := a.config.stores.GetOwn(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractStore(value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) createStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input contract.StoreInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Create(r.Context(), principal, toDomainStoreInput(input))
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractStore(value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *api) updateStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input contract.StoreInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Update(r.Context(), principal, toDomainStoreInput(input))
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractStore(value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) adminStores(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	values, err := a.config.stores.ListForAdmin(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	items, err := toContractStores(values)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.Store]{Items: items})
}

func (a *api) moderateStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input contract.ModerationInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Moderate(r.Context(), principal, r.PathValue("storeId"), string(input.Status), input.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractStore(value)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
