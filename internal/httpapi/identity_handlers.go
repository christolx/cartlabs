package httpapi

import (
	"net/http"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	if a.config.identity == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := a.checkAuthLimit(r, input.Email); err != nil {
		writeError(w, err)
		return
	}
	session, refresh, err := a.config.identity.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, session)
}

func (a *api) demoLogin(w http.ResponseWriter, r *http.Request) {
	if a.config.identity == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	var input struct {
		Role identity.Role `json:"role"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := a.checkAuthLimit(r, string(input.Role)); err != nil {
		writeError(w, err)
		return
	}
	session, refresh, err := a.config.identity.DemoLogin(r.Context(), input.Role)
	if err != nil {
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, session)
}

func (a *api) refresh(w http.ResponseWriter, r *http.Request) {
	if a.config.identity == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	session, refresh, err := a.config.identity.Refresh(r.Context(), a.refreshCookie(r))
	if err != nil {
		a.clearRefreshCookie(w)
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, session)
}

func (a *api) logout(w http.ResponseWriter, r *http.Request) {
	if a.config.identity != nil {
		if err := a.config.identity.Logout(r.Context(), a.refreshCookie(r)); err != nil {
			writeError(w, err)
			return
		}
	}
	a.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) currentUser(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	user, err := a.config.identity.User(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}
