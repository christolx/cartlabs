//go:build integration

package identity

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

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
