package search

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Search(context.Context, string, int) ([]string, error)
	Upsert(context.Context, string, Document) (bool, error)
	Count(context.Context) (int64, error)
	Prune(context.Context, []string, time.Time) (int64, error)
}

type PostgresRepository struct{ pool *pgxpool.Pool }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Search(ctx context.Context, query string, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT product_id::text FROM search_documents
		WHERE name ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%'
		ORDER BY updated_at DESC,product_id DESC LIMIT $2`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query search documents: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan search document: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search documents: %w", err)
	}
	return ids, nil
}

func (r *PostgresRepository) Upsert(ctx context.Context, eventID string, document Document) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin search upsert: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var inserted string
	err = tx.QueryRow(ctx, `INSERT INTO processed_search_events (event_id) VALUES ($1)
		ON CONFLICT DO NOTHING RETURNING event_id::text`, eventID).Scan(&inserted)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("record search event: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO search_documents (product_id,name,description,category_slug,updated_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (product_id) DO UPDATE SET name=EXCLUDED.name,description=EXCLUDED.description,
		category_slug=EXCLUDED.category_slug,updated_at=EXCLUDED.updated_at
		WHERE EXCLUDED.updated_at >= search_documents.updated_at`,
		document.ProductID, document.Name, document.Description, document.CategorySlug, document.UpdatedAt); err != nil {
		return false, fmt.Errorf("upsert search document: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit search upsert: %w", err)
	}
	return true, nil
}

func (r *PostgresRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM search_documents").Scan(&count); err != nil {
		return 0, fmt.Errorf("count search documents: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) Prune(ctx context.Context, keepProductIDs []string, snapshotStartedAt time.Time) (int64, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM search_documents
		WHERE NOT (product_id=ANY($1::uuid[])) AND updated_at <= $2`, keepProductIDs, snapshotStartedAt)
	if err != nil {
		return 0, fmt.Errorf("prune search documents: %w", err)
	}
	return result.RowsAffected(), nil
}
