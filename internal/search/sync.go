package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/christolx/cartlabs/internal/domain"
)

type Upserter interface {
	Upsert(context.Context, string, Document) (bool, error)
}

type SyncHandler struct{ client Upserter }

func NewSyncHandler(client Upserter) *SyncHandler { return &SyncHandler{client: client} }

func (h *SyncHandler) Handle(ctx context.Context, body []byte) error {
	var event UpsertEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return domain.ErrInvalid
	}
	if event.Type != EventType || !validUUID(event.ID) || !validUUID(event.Data.ProductID) {
		return domain.ErrInvalid
	}
	if _, err := h.client.Upsert(ctx, event.ID, event.Data); err != nil {
		return fmt.Errorf("synchronize search document: %w", err)
	}
	return nil
}
