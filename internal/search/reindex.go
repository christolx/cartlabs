package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReindexClient interface {
	Upsert(context.Context, string, Document) (bool, error)
	Prune(context.Context, []string, time.Time) (uint64, error)
}

func Reindex(ctx context.Context, source *pgxpool.Pool, client ReindexClient) (int, uint64, error) {
	snapshotStartedAt := time.Now().UTC()
	documents, err := SourceDocuments(ctx, source)
	if err != nil {
		return 0, 0, err
	}
	keepProductIDs := make([]string, 0, len(documents))
	for _, document := range documents {
		eventID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(document.ProductID+"/reindex/"+snapshotStartedAt.Format(time.RFC3339Nano))).String()
		if _, err := client.Upsert(ctx, eventID, document); err != nil {
			return 0, 0, fmt.Errorf("reindex product %s: %w", document.ProductID, err)
		}
		keepProductIDs = append(keepProductIDs, document.ProductID)
	}
	deleted, err := client.Prune(ctx, keepProductIDs, snapshotStartedAt)
	if err != nil {
		return 0, 0, fmt.Errorf("prune stale search documents: %w", err)
	}
	return len(documents), deleted, nil
}
