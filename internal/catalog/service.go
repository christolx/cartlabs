package catalog

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/google/uuid"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Repository interface {
	Categories(context.Context) ([]Category, error)
	ListForSeller(context.Context, string) ([]Product, error)
	FindForSeller(context.Context, string, string) (Product, error)
	Create(context.Context, string, Product, string) (Product, error)
	Update(context.Context, Product, string, string) (Product, error)
	AddVariant(context.Context, string, Variant, string) (Variant, error)
	AddImage(context.Context, string, ProductImage, string) (ProductImage, error)
	Publish(context.Context, string, string, time.Time) (Product, error)
	Archive(context.Context, string, string, time.Time) (Product, error)
	AdjustInventory(context.Context, string, string, int, string, string, time.Time) (Variant, error)
	ListForAdmin(context.Context) ([]Product, error)
	UpdateStatus(context.Context, string, string, string, string, time.Time) (Product, error)
	ListPublic(context.Context, Filters) (Page, error)
	ListPublicCandidates(context.Context, Filters, []string) (Page, error)
	FindPublic(context.Context, string) (Product, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
	search     CandidateSearcher
	observe    func(string)
}

type CandidateSearcher interface {
	Search(context.Context, string, int) ([]string, error)
}

type Option func(*Service)

func WithCandidateSearcher(search CandidateSearcher) Option {
	return func(service *Service) { service.search = search }
}

func WithSearchObserver(observe func(string)) Option {
	return func(service *Service) { service.observe = observe }
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) Categories(ctx context.Context) ([]Category, error) {
	return s.repository.Categories(ctx)
}

func (s *Service) ListOwn(ctx context.Context, principal identity.Principal) ([]Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListForSeller(ctx, principal.UserID)
}

func (s *Service) FindOwn(ctx context.Context, principal identity.Principal, productID string) (Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return Product{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(productID); err != nil {
		return Product{}, domain.ErrInvalid
	}
	return s.repository.FindForSeller(ctx, principal.UserID, productID)
}

func (s *Service) Create(ctx context.Context, principal identity.Principal, input ProductInput) (Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return Product{}, domain.ErrForbidden
	}
	input, err := validateProduct(input)
	if err != nil {
		return Product{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Product{}, fmt.Errorf("generate product ID: %w", err)
	}
	now := s.now().UTC()
	return s.repository.Create(ctx, principal.UserID, Product{
		ID: id.String(), Category: Category{ID: input.CategoryID}, Name: input.Name, Slug: input.Slug,
		Description: input.Description, Status: "draft", CreatedAt: now, UpdatedAt: now,
	}, principal.UserID)
}

func (s *Service) Update(ctx context.Context, principal identity.Principal, productID string, input ProductInput) (Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return Product{}, domain.ErrForbidden
	}
	input, err := validateProduct(input)
	if err != nil {
		return Product{}, err
	}
	product, err := s.repository.FindForSeller(ctx, principal.UserID, productID)
	if err != nil {
		return Product{}, err
	}
	product.Category.ID, product.Name, product.Slug, product.Description = input.CategoryID, input.Name, input.Slug, input.Description
	product.UpdatedAt = s.now().UTC()
	return s.repository.Update(ctx, product, principal.UserID, principal.UserID)
}

func (s *Service) AddVariant(ctx context.Context, principal identity.Principal, productID string, input VariantInput) (Variant, error) {
	if !principal.Require(identity.RoleSeller) {
		return Variant{}, domain.ErrForbidden
	}
	if _, err := s.repository.FindForSeller(ctx, principal.UserID, productID); err != nil {
		return Variant{}, err
	}
	input, err := validateVariant(input)
	if err != nil {
		return Variant{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Variant{}, fmt.Errorf("generate variant ID: %w", err)
	}
	return s.repository.AddVariant(ctx, productID, Variant{ID: id.String(), SKU: input.SKU, Name: input.Name, Attributes: input.Attributes,
		PriceMinor: input.PriceMinor, Currency: input.Currency, Stock: input.Stock, Active: true}, principal.UserID)
}

func (s *Service) AddImage(ctx context.Context, principal identity.Principal, productID string, input ImageInput) (ProductImage, error) {
	if !principal.Require(identity.RoleSeller) {
		return ProductImage{}, domain.ErrForbidden
	}
	if _, err := s.repository.FindForSeller(ctx, principal.UserID, productID); err != nil {
		return ProductImage{}, err
	}
	input.URL, input.AltText = strings.TrimSpace(input.URL), strings.TrimSpace(input.AltText)
	parsed, err := url.ParseRequestURI(input.URL)
	validLocation := err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https" || parsed.Scheme == "" && strings.HasPrefix(parsed.Path, "/") && !strings.HasPrefix(parsed.Path, "//"))
	if !validLocation || len(input.AltText) > 160 || input.Position < 0 {
		return ProductImage{}, domain.ErrInvalid
	}
	id, err := uuid.NewV7()
	if err != nil {
		return ProductImage{}, fmt.Errorf("generate image ID: %w", err)
	}
	return s.repository.AddImage(ctx, productID, ProductImage{ID: id.String(), URL: input.URL, AltText: input.AltText, Position: input.Position}, principal.UserID)
}

func (s *Service) Publish(ctx context.Context, principal identity.Principal, productID string) (Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return Product{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(productID); err != nil {
		return Product{}, domain.ErrInvalid
	}
	if _, err := s.repository.FindForSeller(ctx, principal.UserID, productID); err != nil {
		return Product{}, err
	}
	return s.repository.Publish(ctx, productID, principal.UserID, s.now().UTC())
}

func (s *Service) Archive(ctx context.Context, principal identity.Principal, productID string) (Product, error) {
	if !principal.Require(identity.RoleSeller) {
		return Product{}, domain.ErrForbidden
	}
	if _, err := uuid.Parse(productID); err != nil {
		return Product{}, domain.ErrInvalid
	}
	if _, err := s.repository.FindForSeller(ctx, principal.UserID, productID); err != nil {
		return Product{}, err
	}
	return s.repository.Archive(ctx, productID, principal.UserID, s.now().UTC())
}

func (s *Service) AdjustInventory(ctx context.Context, principal identity.Principal, variantID string, delta int, reason string) (Variant, error) {
	if !principal.Require(identity.RoleSeller) {
		return Variant{}, domain.ErrForbidden
	}
	reason = strings.TrimSpace(reason)
	if delta == 0 || len(reason) < 1 || len(reason) > 200 {
		return Variant{}, domain.ErrInvalid
	}
	return s.repository.AdjustInventory(ctx, variantID, principal.UserID, delta, reason, principal.UserID, s.now().UTC())
}

func (s *Service) ListForAdmin(ctx context.Context, principal identity.Principal) ([]Product, error) {
	if !principal.Require(identity.RoleAdmin) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListForAdmin(ctx)
}

func (s *Service) UpdateStatus(ctx context.Context, principal identity.Principal, productID, status, reason string) (Product, error) {
	if !principal.Require(identity.RoleAdmin) {
		return Product{}, domain.ErrForbidden
	}
	reason = strings.TrimSpace(reason)
	if _, err := uuid.Parse(productID); err != nil || status != "suspended" && status != "published" || reason == "" || len(reason) > 500 {
		return Product{}, domain.ErrInvalid
	}
	return s.repository.UpdateStatus(ctx, productID, status, reason, principal.UserID, s.now().UTC())
}

func (s *Service) ListPublic(ctx context.Context, filters Filters) (Page, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.CategorySlug = strings.TrimSpace(filters.CategorySlug)
	filters.StoreSlug = strings.ToLower(strings.TrimSpace(filters.StoreSlug))
	if len(filters.Search) > 100 || filters.Page < 1 || filters.PageSize < 1 || filters.PageSize > 100 ||
		filters.StoreSlug != "" && !slugPattern.MatchString(filters.StoreSlug) ||
		filters.MinPrice != nil && *filters.MinPrice < 0 || filters.MaxPrice != nil && *filters.MaxPrice < 0 ||
		filters.MinPrice != nil && filters.MaxPrice != nil && *filters.MinPrice > *filters.MaxPrice {
		return Page{}, domain.ErrInvalid
	}
	if filters.Search != "" && s.search != nil {
		ids, err := s.search.Search(ctx, filters.Search, 2000)
		if err == nil {
			if s.observe != nil {
				s.observe("service")
			}
			return s.repository.ListPublicCandidates(ctx, filters, ids)
		}
		if s.observe != nil {
			s.observe("fallback")
		}
	}
	return s.repository.ListPublic(ctx, filters)
}

func (s *Service) FindPublic(ctx context.Context, slug string) (Product, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return Product{}, domain.ErrInvalid
	}
	return s.repository.FindPublic(ctx, slug)
}

func validateProduct(input ProductInput) (ProductInput, error) {
	input.Name, input.Slug, input.Description = strings.TrimSpace(input.Name), strings.ToLower(strings.TrimSpace(input.Slug)), strings.TrimSpace(input.Description)
	if _, err := uuid.Parse(input.CategoryID); err != nil {
		return ProductInput{}, domain.ErrInvalid
	}
	if len(input.Name) < 2 || len(input.Name) > 160 || !slugPattern.MatchString(input.Slug) || len(input.Description) > 5000 {
		return ProductInput{}, domain.ErrInvalid
	}
	return input, nil
}

func validateVariant(input VariantInput) (VariantInput, error) {
	input.SKU, input.Name, input.Currency = strings.TrimSpace(input.SKU), strings.TrimSpace(input.Name), strings.ToUpper(strings.TrimSpace(input.Currency))
	if input.Attributes == nil {
		input.Attributes = map[string]string{}
	}
	if len(input.SKU) < 2 || len(input.SKU) > 64 || len(input.Name) < 1 || len(input.Name) > 120 || input.Currency != "IDR" || input.PriceMinor < 0 || input.Stock < 0 {
		return VariantInput{}, domain.ErrInvalid
	}
	for key, value := range input.Attributes {
		if strings.TrimSpace(key) == "" || len(key) > 50 || len(value) > 100 {
			return VariantInput{}, domain.ErrInvalid
		}
	}
	return input, nil
}
