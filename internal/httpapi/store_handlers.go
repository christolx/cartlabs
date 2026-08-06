package httpapi

import (
	"net/http"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	marketstore "github.com/christolx/cartlabs/internal/store"
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
	writeJSON(w, http.StatusOK, value)
}

func (a *api) createStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input marketstore.Input
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Create(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (a *api) updateStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input marketstore.Input
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Update(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (a *api) adminStores(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	values, err := a.config.stores.ListForAdmin(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": values})
}

func (a *api) moderateStore(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.stores.Moderate(r.Context(), principal, r.PathValue("storeId"), input.Status, input.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
