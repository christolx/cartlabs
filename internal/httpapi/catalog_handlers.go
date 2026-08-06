package httpapi

import (
	"net/http"
	"strconv"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
)

func (a *api) categories(w http.ResponseWriter, r *http.Request) {
	if a.config.catalog == nil {
		writeError(w, domain.ErrNotFound)
		return
	}
	items, err := a.config.catalog.Categories(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) catalogProducts(w http.ResponseWriter, r *http.Request) {
	filters, err := parseFilters(r)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := a.config.catalog.ListPublic(r.Context(), filters)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func parseFilters(r *http.Request) (catalog.Filters, error) {
	query := r.URL.Query()
	filters := catalog.Filters{Search: query.Get("q"), CategorySlug: query.Get("category"), Page: 1, PageSize: 20}
	var err error
	if value := query.Get("page"); value != "" {
		filters.Page, err = strconv.Atoi(value)
		if err != nil {
			return catalog.Filters{}, domain.ErrInvalid
		}
	}
	if value := query.Get("pageSize"); value != "" {
		filters.PageSize, err = strconv.Atoi(value)
		if err != nil {
			return catalog.Filters{}, domain.ErrInvalid
		}
	}
	if value := query.Get("inStock"); value != "" {
		filters.InStock, err = strconv.ParseBool(value)
		if err != nil {
			return catalog.Filters{}, domain.ErrInvalid
		}
	}
	if filters.MinPrice, err = parseOptionalInt64(query.Get("minPrice")); err != nil {
		return catalog.Filters{}, err
	}
	if filters.MaxPrice, err = parseOptionalInt64(query.Get("maxPrice")); err != nil {
		return catalog.Filters{}, err
	}
	return filters, nil
}

func parseOptionalInt64(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, domain.ErrInvalid
	}
	return &parsed, nil
}

func (a *api) catalogProduct(w http.ResponseWriter, r *http.Request) {
	value, err := a.config.catalog.FindPublic(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (a *api) sellerProducts(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	items, err := a.config.catalog.ListOwn(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) createProduct(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input catalog.ProductInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.Create(r.Context(), principal, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (a *api) updateProduct(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input catalog.ProductInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.Update(r.Context(), principal, r.PathValue("productId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (a *api) createVariant(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input catalog.VariantInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.AddVariant(r.Context(), principal, r.PathValue("productId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (a *api) createImage(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input catalog.ImageInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.AddImage(r.Context(), principal, r.PathValue("productId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

func (a *api) publishProduct(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	value, err := a.config.catalog.Publish(r.Context(), principal, r.PathValue("productId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (a *api) adjustInventory(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input struct {
		Delta  int    `json:"delta"`
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.AdjustInventory(r.Context(), principal, r.PathValue("variantId"), input.Delta, input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

func (a *api) adminProducts(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	items, err := a.config.catalog.ListForAdmin(r.Context(), principal)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *api) moderateProduct(w http.ResponseWriter, r *http.Request, principal identity.Principal) {
	var input struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, err)
		return
	}
	value, err := a.config.catalog.Moderate(r.Context(), principal, r.PathValue("productId"), input.Status, input.Note)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}
