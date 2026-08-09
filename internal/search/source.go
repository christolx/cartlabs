package search

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SourceDocuments(ctx context.Context, pool *pgxpool.Pool) ([]Document, error) {
	rows, err := pool.Query(ctx, `
		SELECT p.id::text,p.name,p.description,c.slug,p.updated_at
		FROM products p JOIN categories c ON c.id=p.category_id
		ORDER BY p.id`)
	if err != nil {
		return nil, fmt.Errorf("query source catalog: %w", err)
	}
	defer rows.Close()
	documents := make([]Document, 0)
	for rows.Next() {
		var document Document
		if err := rows.Scan(&document.ProductID, &document.Name, &document.Description, &document.CategorySlug, &document.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan source catalog: %w", err)
		}
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate source catalog: %w", err)
	}
	return documents, nil
}
