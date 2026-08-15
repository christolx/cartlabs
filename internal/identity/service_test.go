package identity

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
)

type fakeRepository struct {
	users           map[string]User
	sessions        map[string]RefreshSession
	revokedFamilies map[string]bool
}

func (f *fakeRepository) FindByEmail(_ context.Context, email string) (User, error) {
	for _, user := range f.users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, domain.ErrNotFound
}
func (f *fakeRepository) FindByID(_ context.Context, id string) (User, error) {
	user, ok := f.users[id]
	if !ok {
		return User{}, domain.ErrNotFound
	}
	return user, nil
}
func (f *fakeRepository) CreateRefreshSession(_ context.Context, session RefreshSession) error {
	f.sessions[hex.EncodeToString(session.TokenHash)] = session
	return nil
}
func (f *fakeRepository) RotateRefreshSession(_ context.Context, hash []byte, next RefreshSession, now time.Time) (User, error) {
	key := hex.EncodeToString(hash)
	current, ok := f.sessions[key]
	if !ok || f.revokedFamilies[current.FamilyID] || !current.ExpiresAt.After(now) {
		return User{}, domain.ErrUnauthorized
	}
	if current.CreatedAt.IsZero() {
		f.revokedFamilies[current.FamilyID] = true
		return User{}, domain.ErrUnauthorized
	}
	current.CreatedAt = time.Time{}
	f.sessions[key] = current
	next.FamilyID, next.UserID = current.FamilyID, current.UserID
	f.sessions[hex.EncodeToString(next.TokenHash)] = next
	return f.users[current.UserID], nil
}
func (f *fakeRepository) RevokeRefreshFamily(_ context.Context, hash []byte, _ time.Time) error {
	current, ok := f.sessions[hex.EncodeToString(hash)]
	if !ok {
		return domain.ErrNotFound
	}
	f.revokedFamilies[current.FamilyID] = true
	return nil
}
func (f *fakeRepository) ListUsers(context.Context) ([]AdminUser, error) {
	users := make([]AdminUser, 0, len(f.users))
	for _, user := range f.users {
		users = append(users, AdminUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: user.Role, Status: user.Status, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt})
	}
	return users, nil
}
func (f *fakeRepository) UpdateUserStatus(_ context.Context, actorID, userID, status, _ string, now time.Time) (AdminUser, error) {
	user, ok := f.users[userID]
	if !ok {
		return AdminUser{}, domain.ErrNotFound
	}
	if user.Status == status || actorID == userID && status == "suspended" {
		return AdminUser{}, domain.ErrConflict
	}
	if user.Role == RoleAdmin && status == "suspended" {
		active := 0
		for _, candidate := range f.users {
			if candidate.Role == RoleAdmin && candidate.Status == "active" {
				active++
			}
		}
		if active <= 1 {
			return AdminUser{}, domain.ErrConflict
		}
	}
	user.Status, user.UpdatedAt = status, now
	f.users[userID] = user
	return AdminUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: user.Role, Status: user.Status, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}, nil
}

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "correct-horse-battery") {
		t.Fatal("valid password rejected")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("invalid password accepted")
	}
	if VerifyPassword("$argon2id$v=19$m=999999999,t=2,p=2$bad$bad", "password") {
		t.Fatal("unsafe hash parameters accepted")
	}
}

func TestRefreshRotationAndReplayRevokesFamily(t *testing.T) {
	now := time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)
	hash, err := HashPassword("demo-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeRepository{
		users:    map[string]User{"u1": {ID: "u1", Email: "buyer@example.com", PasswordHash: hash, DisplayName: "Buyer", Role: RoleBuyer, Status: "active"}},
		sessions: map[string]RefreshSession{}, revokedFamilies: map[string]bool{},
	}
	service, err := NewService(repository, Config{AccessSecret: []byte("01234567890123456789012345678901"), AccessTTL: 15 * time.Minute, RefreshTTL: time.Hour, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	session, first, err := service.Login(context.Background(), " BUYER@example.com ", "demo-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	if session.User.ID != "u1" || first == "" {
		t.Fatalf("unexpected session: %#v", session)
	}
	principal, err := service.Authenticate(context.Background(), session.AccessToken)
	if err != nil || principal.UserID != "u1" {
		t.Fatalf("authenticate = %#v, %v", principal, err)
	}
	_, second, err := service.Refresh(context.Background(), first)
	if err != nil || second == first {
		t.Fatalf("rotation failed: %v", err)
	}
	if _, _, err := service.Refresh(context.Background(), first); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("replay error = %v", err)
	}
	if _, _, err := service.Refresh(context.Background(), second); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("family remained usable after replay: %v", err)
	}
}

func TestDemoLoginRequiresExplicitMode(t *testing.T) {
	repository := &fakeRepository{users: map[string]User{}, sessions: map[string]RefreshSession{}, revokedFamilies: map[string]bool{}}
	service, err := NewService(repository, Config{AccessSecret: []byte("01234567890123456789012345678901"), AccessTTL: time.Minute, RefreshTTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.DemoLogin(context.Background(), RoleAdmin); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("error = %v", err)
	}
}

func TestRateLimitKeyDoesNotExposeSubject(t *testing.T) {
	key := RateLimitKey("127.0.0.1:1234", "Buyer@Example.com")
	if len(key) != 32 {
		t.Fatalf("key length = %d", len(key))
	}
	if key == RateLimitKey("127.0.0.1:1234", "seller@example.com") {
		t.Fatal("different subjects generated same key")
	}
}

func TestAdminUserStatusValidationAndSafety(t *testing.T) {
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{users: map[string]User{
		"01989f00-0000-7000-8000-000000000001": {ID: "01989f00-0000-7000-8000-000000000001", Role: RoleAdmin, Status: "active"},
		"01989f00-0000-7000-8000-000000000002": {ID: "01989f00-0000-7000-8000-000000000002", Role: RoleBuyer, Status: "active"},
	}, sessions: map[string]RefreshSession{}, revokedFamilies: map[string]bool{}}
	service, err := NewService(repository, Config{AccessSecret: []byte("01234567890123456789012345678901"), AccessTTL: time.Minute, RefreshTTL: time.Hour, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	admin := Principal{UserID: "01989f00-0000-7000-8000-000000000001", Role: RoleAdmin}
	otherAdmin := Principal{UserID: "01989f00-0000-7000-8000-000000000099", Role: RoleAdmin}
	buyer := Principal{UserID: "01989f00-0000-7000-8000-000000000002", Role: RoleBuyer}
	tests := []struct {
		name, id, status, reason string
		principal                Principal
		want                     error
	}{
		{"non-admin", buyer.UserID, "suspended", "policy", buyer, domain.ErrForbidden},
		{"bad id", "bad", "suspended", "policy", admin, domain.ErrInvalid},
		{"bad status", buyer.UserID, "deleted", "policy", admin, domain.ErrInvalid},
		{"blank reason", buyer.UserID, "suspended", "   ", admin, domain.ErrInvalid},
		{"self suspension", admin.UserID, "suspended", "policy", admin, domain.ErrConflict},
		{"last active admin", admin.UserID, "suspended", "policy", otherAdmin, domain.ErrConflict},
	}
	if _, err := service.ListUsers(context.Background(), buyer); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("buyer list error = %v", err)
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.UpdateUserStatus(context.Background(), test.principal, test.id, test.status, test.reason)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	updated, err := service.UpdateUserStatus(context.Background(), admin, buyer.UserID, " suspended ", " policy violation ")
	if err != nil || updated.Status != "suspended" || !updated.UpdatedAt.Equal(now) {
		t.Fatalf("updated = %#v, err = %v", updated, err)
	}
	if _, err := service.UpdateUserStatus(context.Background(), admin, buyer.UserID, "suspended", "again"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("redundant transition error = %v", err)
	}
}

func TestStatusChangeInvalidatesOldAccessUntilFreshLogin(t *testing.T) {
	now := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	passwordHash, err := HashPassword("demo-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	adminID := "01989f00-0000-7000-8000-000000000001"
	buyerID := "01989f00-0000-7000-8000-000000000002"
	repository := &fakeRepository{users: map[string]User{
		adminID: {ID: adminID, Email: "admin@example.com", Role: RoleAdmin, Status: "active", UpdatedAt: now},
		buyerID: {ID: buyerID, Email: "buyer@example.com", PasswordHash: passwordHash, Role: RoleBuyer, Status: "active", UpdatedAt: now},
	}, sessions: map[string]RefreshSession{}, revokedFamilies: map[string]bool{}}
	service, err := NewService(repository, Config{AccessSecret: []byte("01234567890123456789012345678901"), AccessTTL: time.Hour, RefreshTTL: time.Hour, Now: func() time.Time { return now }})
	if err != nil {
		t.Fatal(err)
	}
	session, _, err := service.Login(context.Background(), "buyer@example.com", "demo-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	admin := Principal{UserID: adminID, Role: RoleAdmin}
	now = now.Add(time.Second)
	if _, err := service.UpdateUserStatus(context.Background(), admin, buyerID, "suspended", "policy"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	if _, err := service.UpdateUserStatus(context.Background(), admin, buyerID, "active", "appeal"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), session.AccessToken); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("old access error = %v", err)
	}
	fresh, _, err := service.Login(context.Background(), "buyer@example.com", "demo-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Authenticate(context.Background(), fresh.AccessToken); err != nil {
		t.Fatalf("fresh access error = %v", err)
	}
}
