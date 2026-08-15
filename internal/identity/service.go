package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const dummyPasswordHash = "$argon2id$v=19$m=65536,t=2,p=2$Dc5YGbStWBSDW1FlEUTakw$5LJWmV9gPuIrgqLXYWBmusWCPL+jArlcW9eFaccK+/g"

type Repository interface {
	FindByEmail(context.Context, string) (User, error)
	FindByID(context.Context, string) (User, error)
	CreateRefreshSession(context.Context, RefreshSession) error
	RotateRefreshSession(context.Context, []byte, RefreshSession, time.Time) (User, error)
	RevokeRefreshFamily(context.Context, []byte, time.Time) error
	ListUsers(context.Context) ([]AdminUser, error)
	UpdateUserStatus(context.Context, string, string, string, string, time.Time) (AdminUser, error)
}

type Config struct {
	AccessSecret []byte
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	DemoMode     bool
	Now          func() time.Time
}

type Service struct {
	repository Repository
	tokens     tokenManager
	refreshTTL time.Duration
	demoMode   bool
	now        func() time.Time
}

func NewService(repository Repository, cfg Config) (*Service, error) {
	if len(cfg.AccessSecret) < 32 {
		return nil, fmt.Errorf("access token secret must contain at least 32 bytes")
	}
	if cfg.AccessTTL <= 0 || cfg.RefreshTTL <= 0 {
		return nil, fmt.Errorf("token TTL must be positive")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Service{
		repository: repository,
		tokens:     tokenManager{secret: cfg.AccessSecret, issuer: "cartlabs", audience: "cartlabs-web", ttl: cfg.AccessTTL, now: cfg.Now},
		refreshTTL: cfg.RefreshTTL,
		demoMode:   cfg.DemoMode,
		now:        cfg.Now,
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, string, error) {
	user, err := s.repository.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	hash := user.PasswordHash
	if err != nil {
		hash = dummyPasswordHash
	}
	passwordValid := VerifyPassword(hash, password)
	if errors.Is(err, domain.ErrNotFound) || err == nil && (user.Status != "active" || !passwordValid) {
		return Session{}, "", domain.ErrUnauthorized
	}
	if err != nil {
		return Session{}, "", err
	}
	return s.createSession(ctx, user)
}

func (s *Service) DemoLogin(ctx context.Context, role Role) (Session, string, error) {
	if !s.demoMode {
		return Session{}, "", domain.ErrForbidden
	}
	emails := map[Role]string{
		RoleBuyer: "buyer@demo.cartlabs.local", RoleSeller: "seller@demo.cartlabs.local", RoleAdmin: "admin@demo.cartlabs.local",
	}
	email, ok := emails[role]
	if !ok {
		return Session{}, "", domain.ErrInvalid
	}
	user, err := s.repository.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) || err == nil && (user.Status != "active" || user.Role != role) {
		return Session{}, "", domain.ErrUnauthorized
	}
	if err != nil {
		return Session{}, "", err
	}
	return s.createSession(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, raw string) (Session, string, error) {
	if raw == "" {
		return Session{}, "", domain.ErrUnauthorized
	}
	nextRaw, nextHash, err := newOpaqueToken()
	if err != nil {
		return Session{}, "", err
	}
	now := s.now().UTC()
	nextID, err := uuid.NewV7()
	if err != nil {
		return Session{}, "", fmt.Errorf("generate session ID: %w", err)
	}
	next := RefreshSession{ID: nextID.String(), TokenHash: nextHash, ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now}
	user, err := s.repository.RotateRefreshSession(ctx, hashOpaqueToken(raw), next, now)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) || errors.Is(err, domain.ErrNotFound) {
			return Session{}, "", domain.ErrUnauthorized
		}
		return Session{}, "", err
	}
	access, expiresIn, err := s.tokens.issue(user)
	if err != nil {
		return Session{}, "", err
	}
	return Session{AccessToken: access, TokenType: "Bearer", ExpiresIn: expiresIn, User: user}, nextRaw, nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	if err := s.repository.RevokeRefreshFamily(ctx, hashOpaqueToken(raw), s.now().UTC()); err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	return nil
}

func (s *Service) Authenticate(ctx context.Context, raw string) (Principal, error) {
	principal, err := s.tokens.parse(raw)
	if err != nil {
		return Principal{}, domain.ErrUnauthorized
	}
	user, err := s.repository.FindByID(ctx, principal.UserID)
	if errors.Is(err, domain.ErrNotFound) || err == nil && (user.Status != "active" || user.Role != principal.Role || user.UpdatedAt.UnixNano() != principal.UserVersion) {
		return Principal{}, domain.ErrUnauthorized
	}
	if err != nil {
		return Principal{}, err
	}
	return principal, nil
}

func (s *Service) User(ctx context.Context, principal Principal) (User, error) {
	return s.repository.FindByID(ctx, principal.UserID)
}

func (s *Service) ListUsers(ctx context.Context, principal Principal) ([]AdminUser, error) {
	if !principal.Require(RoleAdmin) {
		return nil, domain.ErrForbidden
	}
	return s.repository.ListUsers(ctx)
}

func (s *Service) UpdateUserStatus(ctx context.Context, principal Principal, userID, status, reason string) (AdminUser, error) {
	if !principal.Require(RoleAdmin) {
		return AdminUser{}, domain.ErrForbidden
	}
	status = strings.ToLower(strings.TrimSpace(status))
	reason = strings.TrimSpace(reason)
	if _, err := uuid.Parse(userID); err != nil || status != "active" && status != "suspended" || reason == "" || len(reason) > 500 {
		return AdminUser{}, domain.ErrInvalid
	}
	return s.repository.UpdateUserStatus(ctx, principal.UserID, userID, status, reason, s.now().UTC())
}

func (s *Service) createSession(ctx context.Context, user User) (Session, string, error) {
	raw, hash, err := newOpaqueToken()
	if err != nil {
		return Session{}, "", err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Session{}, "", fmt.Errorf("generate session ID: %w", err)
	}
	familyID, err := uuid.NewV7()
	if err != nil {
		return Session{}, "", fmt.Errorf("generate session family ID: %w", err)
	}
	now := s.now().UTC()
	if err := s.repository.CreateRefreshSession(ctx, RefreshSession{
		ID: id.String(), FamilyID: familyID.String(), UserID: user.ID, TokenHash: hash,
		ExpiresAt: now.Add(s.refreshTTL), CreatedAt: now,
	}); err != nil {
		return Session{}, "", err
	}
	access, expiresIn, err := s.tokens.issue(user)
	if err != nil {
		return Session{}, "", err
	}
	return Session{AccessToken: access, TokenType: "Bearer", ExpiresIn: expiresIn, User: user}, raw, nil
}
