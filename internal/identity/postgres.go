package identity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(r.pool.QueryRow(ctx, `
		SELECT id::text, email, password_hash, display_name, role::text, status::text, created_at, updated_at
		FROM users WHERE email = $1`, email))
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (User, error) {
	return scanUser(r.pool.QueryRow(ctx, `
		SELECT id::text, email, password_hash, display_name, role::text, status::text, created_at, updated_at
		FROM users WHERE id = $1`, id))
}

type rowScanner interface {
	Scan(...any) error
}

func scanUser(row rowScanner) (User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, domain.ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) CreateRefreshSession(ctx context.Context, session RefreshSession) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin refresh session creation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockIdentityUser(ctx, tx, session.UserID); err != nil {
		return err
	}
	var status string
	if err := tx.QueryRow(ctx, `SELECT status::text FROM users WHERE id=$1`, session.UserID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUnauthorized
		}
		return fmt.Errorf("check refresh session user: %w", err)
	}
	if status != "active" {
		return domain.ErrUnauthorized
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO refresh_sessions (id, family_id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		session.ID, session.FamilyID, session.UserID, session.TokenHash, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("create refresh session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit refresh session creation: %w", err)
	}
	return nil
}

func (r *PostgresRepository) RotateRefreshSession(ctx context.Context, currentHash []byte, next RefreshSession, now time.Time) (User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin refresh rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	if err := tx.QueryRow(ctx, `SELECT user_id::text FROM refresh_sessions WHERE token_hash=$1`, currentHash).Scan(&userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, domain.ErrNotFound
		}
		return User{}, fmt.Errorf("find refresh session user: %w", err)
	}
	if err := lockIdentityUser(ctx, tx, userID); err != nil {
		return User{}, err
	}

	var current RefreshSession
	var rotatedAt, revokedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT id::text, family_id::text, user_id::text, expires_at, rotated_at, revoked_at
		FROM refresh_sessions WHERE token_hash = $1 FOR UPDATE`, currentHash).
		Scan(&current.ID, &current.FamilyID, &current.UserID, &current.ExpiresAt, &rotatedAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, domain.ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("lock refresh session: %w", err)
	}

	if rotatedAt != nil || revokedAt != nil || !current.ExpiresAt.After(now) {
		if _, err := tx.Exec(ctx, `
			UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, $2)
			WHERE family_id = $1`, current.FamilyID, now); err != nil {
			return User{}, fmt.Errorf("revoke replayed refresh family: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return User{}, fmt.Errorf("commit replay revocation: %w", err)
		}
		return User{}, domain.ErrUnauthorized
	}

	if _, err := tx.Exec(ctx, `UPDATE refresh_sessions SET rotated_at = $2 WHERE id = $1`, current.ID, now); err != nil {
		return User{}, fmt.Errorf("mark refresh session rotated: %w", err)
	}
	next.FamilyID = current.FamilyID
	next.UserID = current.UserID
	if _, err := tx.Exec(ctx, `
		INSERT INTO refresh_sessions (id, family_id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		next.ID, next.FamilyID, next.UserID, next.TokenHash, next.ExpiresAt, next.CreatedAt); err != nil {
		return User{}, fmt.Errorf("insert rotated refresh session: %w", err)
	}
	user, err := scanUser(tx.QueryRow(ctx, `
		SELECT id::text, email, password_hash, display_name, role::text, status::text, created_at, updated_at
		FROM users WHERE id = $1`, current.UserID))
	if err != nil || user.Status != "active" {
		return User{}, domain.ErrUnauthorized
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit refresh rotation: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) RevokeRefreshFamily(ctx context.Context, tokenHash []byte, now time.Time) error {
	command, err := r.pool.Exec(ctx, `
		UPDATE refresh_sessions SET revoked_at = COALESCE(revoked_at, $2)
		WHERE family_id = (SELECT family_id FROM refresh_sessions WHERE token_hash = $1)`, tokenHash, now)
	if err != nil {
		return fmt.Errorf("revoke refresh family: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) ListUsers(ctx context.Context) ([]AdminUser, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id::text,email,display_name,role::text,status::text,created_at,updated_at
		FROM users ORDER BY created_at DESC,id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := make([]AdminUser, 0)
	for rows.Next() {
		user, err := scanAdminUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func scanAdminUser(row rowScanner) (AdminUser, error) {
	var user AdminUser
	if err := row.Scan(&user.ID, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminUser{}, domain.ErrNotFound
		}
		return AdminUser{}, fmt.Errorf("scan admin user: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) UpdateUserStatus(ctx context.Context, actorID, userID, status, reason string, now time.Time) (AdminUser, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AdminUser{}, fmt.Errorf("begin user status update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// Serialize status changes so concurrent admin suspensions cannot bypass last-admin protection.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(1128352842)`); err != nil {
		return AdminUser{}, fmt.Errorf("lock user status updates: %w", err)
	}
	if err := lockIdentityUser(ctx, tx, userID); err != nil {
		return AdminUser{}, err
	}
	current, err := scanAdminUser(tx.QueryRow(ctx, `
		SELECT id::text,email,display_name,role::text,status::text,created_at,updated_at
		FROM users WHERE id=$1 FOR UPDATE`, userID))
	if err != nil {
		return AdminUser{}, err
	}
	if current.Status == status || actorID == userID && status == "suspended" {
		return AdminUser{}, domain.ErrConflict
	}
	if current.Role == RoleAdmin && status == "suspended" {
		var activeAdmins int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='admin' AND status='active'`).Scan(&activeAdmins); err != nil {
			return AdminUser{}, fmt.Errorf("count active admins: %w", err)
		}
		if activeAdmins <= 1 {
			return AdminUser{}, domain.ErrConflict
		}
	}
	updated, err := scanAdminUser(tx.QueryRow(ctx, `
		UPDATE users SET status=$2,updated_at=$3 WHERE id=$1
		RETURNING id::text,email,display_name,role::text,status::text,created_at,updated_at`, userID, status, now))
	if err != nil {
		return AdminUser{}, err
	}
	if status == "suspended" {
		if _, err := tx.Exec(ctx, `UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE user_id=$1`, userID, now); err != nil {
			return AdminUser{}, fmt.Errorf("revoke user refresh sessions: %w", err)
		}
	}
	metadata, err := json.Marshal(map[string]string{"from": current.Status, "to": status, "reason": reason})
	if err != nil {
		return AdminUser{}, fmt.Errorf("encode user status audit: %w", err)
	}
	auditID, err := uuid.NewV7()
	if err != nil {
		return AdminUser{}, fmt.Errorf("generate audit ID: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_log (id,actor_id,action,resource_type,resource_id,metadata,created_at)
		VALUES ($1,$2,$3,'user',$4,$5,$6)`, auditID.String(), actorID, "user.status."+status, userID, metadata, now); err != nil {
		return AdminUser{}, fmt.Errorf("insert user status audit: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return AdminUser{}, fmt.Errorf("commit user status update: %w", err)
	}
	return updated, nil
}

func lockIdentityUser(ctx context.Context, tx pgx.Tx, userID string) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('identity.user:' || $1, 0))`, userID); err != nil {
		return fmt.Errorf("lock identity user: %w", err)
	}
	return nil
}
