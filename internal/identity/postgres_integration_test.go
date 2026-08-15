//go:build integration

package identity

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUserStatusUpdateIsAtomicAndRevokesSessions(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	actorID, targetID, sessionID, familyID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_log WHERE actor_id=$1 OR resource_id=$2`, actorID, targetID)
		_, _ = pool.Exec(ctx, `DELETE FROM refresh_sessions WHERE user_id=$1`, targetID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, actorID, targetID)
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES
		($1,$2,'hash','Integration Admin','admin','active'),($3,$4,'hash','Integration Buyer','buyer','active')`,
		actorID, actorID+"@example.com", targetID, targetID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO refresh_sessions (id,family_id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4,$5)`,
		sessionID, familyID, targetID, []byte("integration-session-1"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	repository := NewPostgresRepository(pool)
	updated, err := repository.UpdateUserStatus(ctx, actorID, targetID, "suspended", "integration policy", now)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "suspended" || !updated.UpdatedAt.Equal(now) {
		t.Fatalf("updated = %#v", updated)
	}
	var revokedAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT revoked_at FROM refresh_sessions WHERE id=$1`, sessionID).Scan(&revokedAt); err != nil || revokedAt == nil {
		t.Fatalf("revokedAt=%v err=%v", revokedAt, err)
	}
	if err := repository.CreateRefreshSession(ctx, RefreshSession{
		ID: uuid.NewString(), FamilyID: uuid.NewString(), UserID: targetID, TokenHash: []byte("suspended-session"),
		ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now().UTC(),
	}); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("suspended user session creation error = %v", err)
	}
	var action string
	var raw []byte
	if err := pool.QueryRow(ctx, `SELECT action,metadata FROM audit_log WHERE actor_id=$1 AND resource_id=$2 ORDER BY created_at DESC LIMIT 1`, actorID, targetID).Scan(&action, &raw); err != nil {
		t.Fatal(err)
	}
	var metadata map[string]string
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	if action != "user.status.suspended" || metadata["from"] != "active" || metadata["to"] != "suspended" || metadata["reason"] != "integration policy" {
		t.Fatalf("action=%q metadata=%v", action, metadata)
	}

	if _, err := repository.UpdateUserStatus(ctx, actorID, targetID, "active", "appeal accepted", now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	secondSessionID := uuid.NewString()
	if _, err := pool.Exec(ctx, `INSERT INTO refresh_sessions (id,family_id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4,$5)`,
		secondSessionID, uuid.NewString(), targetID, []byte("integration-session-2"), time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.UpdateUserStatus(ctx, uuid.NewString(), targetID, "suspended", "must roll back", now.Add(2*time.Second)); err == nil {
		t.Fatal("status update with missing audit actor succeeded")
	}
	var status string
	var secondRevokedAt *time.Time
	if err := pool.QueryRow(ctx, `SELECT status::text FROM users WHERE id=$1`, targetID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT revoked_at FROM refresh_sessions WHERE id=$1`, secondSessionID).Scan(&secondRevokedAt); err != nil {
		t.Fatal(err)
	}
	if status != "active" || secondRevokedAt != nil {
		t.Fatalf("failed transaction leaked status=%q revokedAt=%v", status, secondRevokedAt)
	}
}

func TestConcurrentSuspensionWinsOverRefreshRotation(t *testing.T) {
	databaseURL := os.Getenv("INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("INTEGRATION_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	actorID, targetID := uuid.NewString(), uuid.NewString()
	sessionID, familyID := uuid.NewString(), uuid.NewString()
	currentHash := []byte("concurrent-refresh-current")
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM audit_log WHERE actor_id=$1 OR resource_id=$2`, actorID, targetID)
		_, _ = pool.Exec(ctx, `DELETE FROM refresh_sessions WHERE user_id=$1`, targetID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, actorID, targetID)
	}()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id,email,password_hash,display_name,role,status) VALUES
		($1,$2,'hash','Concurrent Admin','admin','active'),($3,$4,'hash','Concurrent Buyer','buyer','active')`,
		actorID, actorID+"@example.com", targetID, targetID+"@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO refresh_sessions (id,family_id,user_id,token_hash,expires_at) VALUES ($1,$2,$3,$4,$5)`,
		sessionID, familyID, targetID, currentHash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	gate, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gate.Rollback(ctx) }()
	if err := lockIdentityUser(ctx, gate, targetID); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgresRepository(pool)
	statusResult := make(chan error, 1)
	go func() {
		_, err := repository.UpdateUserStatus(ctx, actorID, targetID, "suspended", "concurrent policy", time.Now().UTC())
		statusResult <- err
	}()

	probe, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Release()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var acquired bool
		if err := probe.QueryRow(ctx, `SELECT pg_try_advisory_lock(1128352842)`).Scan(&acquired); err != nil {
			t.Fatal(err)
		}
		if !acquired {
			break
		}
		if _, err := probe.Exec(ctx, `SELECT pg_advisory_unlock(1128352842)`); err != nil {
			t.Fatal(err)
		}
		if time.Now().After(deadline) {
			t.Fatal("status update did not reach identity lock")
		}
		time.Sleep(10 * time.Millisecond)
	}

	rotateResult := make(chan error, 1)
	go func() {
		_, err := repository.RotateRefreshSession(ctx, currentHash, RefreshSession{
			ID: uuid.NewString(), TokenHash: []byte("concurrent-refresh-next"), ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now().UTC(),
		}, time.Now().UTC())
		rotateResult <- err
	}()
	if err := gate.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var statusErr error
	select {
	case statusErr = <-statusResult:
	case <-time.After(5 * time.Second):
		t.Fatal("status update timed out")
	}
	if statusErr != nil {
		t.Fatalf("status update: %v", statusErr)
	}
	var rotateErr error
	select {
	case rotateErr = <-rotateResult:
	case <-time.After(5 * time.Second):
		t.Fatal("refresh rotation timed out")
	}
	if !errors.Is(rotateErr, domain.ErrUnauthorized) {
		t.Fatalf("refresh rotation error = %v", rotateErr)
	}
	var activeSessions int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM refresh_sessions WHERE user_id=$1 AND revoked_at IS NULL`, targetID).Scan(&activeSessions); err != nil {
		t.Fatal(err)
	}
	if activeSessions != 0 {
		t.Fatalf("active sessions after suspension = %d", activeSessions)
	}
}
