package httpapi

import (
	"net/http"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

func (a *api) login(w http.ResponseWriter, r *http.Request) {
	if a.config.identity == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	var input contract.LoginRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := a.checkAuthLimit(r, string(input.Email)); err != nil {
		writeError(w, err)
		return
	}
	session, refresh, err := a.config.identity.Login(r.Context(), string(input.Email), input.Password)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractSession(session)
	if err != nil {
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, response)
}

func (a *api) demoLogin(w http.ResponseWriter, r *http.Request) {
	if a.config.identity == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	var input contract.DemoLoginRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	role, err := toDomainRole(input.Role)
	if err != nil {
		writeError(w, domain.ErrInvalid)
		return
	}
	if err := a.checkAuthLimit(r, string(role)); err != nil {
		writeError(w, err)
		return
	}
	session, refresh, err := a.config.identity.DemoLogin(r.Context(), role)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractSession(session)
	if err != nil {
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractSession(session)
	if err != nil {
		writeError(w, err)
		return
	}
	a.setRefreshCookie(w, refresh)
	writeJSON(w, http.StatusOK, response)
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
	response, err := toContractUser(user)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *api) adminUsers(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	users, err := a.config.identity.ListUsers(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractAdminUsers(users)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, itemsResponse[contract.AdminUser]{Items: response})
}

func (a *api) updateUserStatus(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input contract.UserStatusInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	user, err := a.config.identity.UpdateUserStatus(r.Context(), principal, r.PathValue("userId"), string(input.Status), input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	response, err := toContractAdminUser(user)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}
