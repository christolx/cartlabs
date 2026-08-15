package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/christolx/cartlabs/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

type rowScanner interface{ Scan(...any) error }

func scanStore(row rowScanner) (Store, error) {
	var result Store
	err := row.Scan(&result.ID, &result.SellerID, &result.Name, &result.Slug, &result.Description,
		&result.Status, &result.ModerationNote, &result.CreatedAt, &result.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Store{}, domain.ErrNotFound
	}
	if err != nil {
		return Store{}, fmt.Errorf("scan store: %w", err)
	}
	return result, nil
}

const storeColumns = `id::text, seller_id::text, name, slug, description, status::text, moderation_note, created_at, updated_at`

func (r *PostgresRepository) FindBySeller(ctx context.Context, sellerID string) (Store, error) {
	return scanStore(r.pool.QueryRow(ctx, `SELECT `+storeColumns+` FROM stores WHERE seller_id = $1`, sellerID))
}

func (r *PostgresRepository) FindPublicBySlug(ctx context.Context, slug string) (Profile, error) {
	var profile Profile
	err := r.pool.QueryRow(ctx, `
		SELECT s.id::text,s.name,s.slug,s.description,u.display_name,s.created_at
		FROM stores s JOIN users u ON u.id=s.seller_id
		WHERE s.slug=$1 AND s.status='approved'`, slug).
		Scan(&profile.ID, &profile.Name, &profile.Slug, &profile.Description, &profile.SellerDisplayName, &profile.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, domain.ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("find public store: %w", err)
	}
	return profile, nil
}

func (r *PostgresRepository) Create(ctx context.Context, value Store, actorID string) (Store, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Store{}, fmt.Errorf("begin create store: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	created, err := scanStore(tx.QueryRow(ctx, `
		INSERT INTO stores (id, seller_id, name, slug, description, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$7) RETURNING `+storeColumns,
		value.ID, value.SellerID, value.Name, value.Slug, value.Description, value.Status, value.CreatedAt))
	if err != nil {
		return Store{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "store.created", "store", created.ID, created.CreatedAt); err != nil {
		return Store{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Store{}, fmt.Errorf("commit create store: %w", err)
	}
	return created, nil
}

func (r *PostgresRepository) Update(ctx context.Context, value Store, actorID string) (Store, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Store{}, fmt.Errorf("begin update store: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanStore(tx.QueryRow(ctx, `
		UPDATE stores SET name=$2,slug=$3,description=$4,status='pending',moderation_note='',updated_at=$5
		WHERE id=$1 AND seller_id=$6 RETURNING `+storeColumns,
		value.ID, value.Name, value.Slug, value.Description, value.UpdatedAt, value.SellerID))
	if err != nil {
		return Store{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "store.updated", "store", updated.ID, updated.UpdatedAt); err != nil {
		return Store{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Store{}, fmt.Errorf("commit update store: %w", err)
	}
	return updated, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]Store, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+storeColumns+` FROM stores ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	defer rows.Close()
	result := make([]Store, 0)
	for rows.Next() {
		value, err := scanStore(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Moderate(ctx context.Context, id, status, note, actorID string, now time.Time) (Store, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Store{}, fmt.Errorf("begin moderate store: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	updated, err := scanStore(tx.QueryRow(ctx, `
		UPDATE stores SET status=$2,moderation_note=$3,updated_at=$4 WHERE id=$1 RETURNING `+storeColumns,
		id, status, note, now))
	if err != nil {
		return Store{}, mapWriteError(err)
	}
	if err := insertAudit(ctx, tx, actorID, "store.moderated."+status, "store", id, now); err != nil {
		return Store{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Store{}, fmt.Errorf("commit moderate store: %w", err)
	}
	return updated, nil
}

type auditExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertAudit(ctx context.Context, tx auditExecutor, actorID, action, resourceType, resourceID string, now time.Time) error {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate audit ID: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO audit_log (id,actor_id,action,resource_type,resource_id,created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		id.String(), actorID, action, resourceType, resourceID, now); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func mapWriteError(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return err
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}
