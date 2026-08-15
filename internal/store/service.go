package store

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/christolx/cartlabs/internal/identity"
	"github.com/google/uuid"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Repository interface {
	FindBySeller(context.Context, string) (Store, error)
	Create(context.Context, Store, string) (Store, error)
	Update(context.Context, Store, string) (Store, error)
	List(context.Context) ([]Store, error)
	Moderate(context.Context, string, string, string, string, time.Time) (Store, error)
	FindPublicBySlug(context.Context, string) (Profile, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) GetOwn(ctx context.Context, principal identity.Principal) (Store, error) {
	if !principal.Require(identity.RoleSeller) {
		return Store{}, domain.ErrForbidden
	}
	return s.repository.FindBySeller(ctx, principal.UserID)
}

func (s *Service) FindPublic(ctx context.Context, slug string) (Profile, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if !slugPattern.MatchString(slug) {
		return Profile{}, domain.ErrInvalid
	}
	return s.repository.FindPublicBySlug(ctx, slug)
}

func (s *Service) Create(ctx context.Context, principal identity.Principal, input Input) (Store, error) {
	if !principal.Require(identity.RoleSeller) {
		return Store{}, domain.ErrForbidden
	}
	input, err := validateInput(input)
	if err != nil {
		return Store{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Store{}, fmt.Errorf("generate store ID: %w", err)
	}
	now := s.now().UTC()
	return s.repository.Create(ctx, Store{
		ID: id.String(), SellerID: principal.UserID, Name: input.Name, Slug: input.Slug,
		Description: input.Description, Status: "pending", CreatedAt: now, UpdatedAt: now,
	}, principal.UserID)
}

func (s *Service) Update(ctx context.Context, principal identity.Principal, input Input) (Store, error) {
	if !principal.Require(identity.RoleSeller) {
		return Store{}, domain.ErrForbidden
	}
	input, err := validateInput(input)
	if err != nil {
		return Store{}, err
	}
	current, err := s.repository.FindBySeller(ctx, principal.UserID)
	if err != nil {
		return Store{}, err
	}
	current.Name = input.Name
	current.Slug = input.Slug
	current.Description = input.Description
	current.Status = "pending"
	current.ModerationNote = ""
	current.UpdatedAt = s.now().UTC()
	return s.repository.Update(ctx, current, principal.UserID)
}

func (s *Service) ListForAdmin(ctx context.Context, principal identity.Principal) ([]Store, error) {
	if !principal.Require(identity.RoleAdmin) {
		return nil, domain.ErrForbidden
	}
	return s.repository.List(ctx)
}

func (s *Service) Moderate(ctx context.Context, principal identity.Principal, id, status, note string) (Store, error) {
	if !principal.Require(identity.RoleAdmin) {
		return Store{}, domain.ErrForbidden
	}
	if status != "approved" && status != "rejected" {
		return Store{}, domain.ErrInvalid
	}
	if len(note) > 500 {
		return Store{}, domain.ErrInvalid
	}
	return s.repository.Moderate(ctx, id, status, strings.TrimSpace(note), principal.UserID, s.now().UTC())
}

func validateInput(input Input) (Input, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(strings.ToLower(input.Slug))
	input.Description = strings.TrimSpace(input.Description)
	if len(input.Name) < 2 || len(input.Name) > 100 || !slugPattern.MatchString(input.Slug) || len(input.Description) > 1000 {
		return Input{}, domain.ErrInvalid
	}
	return input, nil
}
