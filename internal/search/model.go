package search

import "time"

const (
	EventType       = "catalog.search.upsert.v1"
	DefaultLimit    = 2000
	MaximumRPCItems = 5000
)

type Document struct {
	ProductID    string    `json:"productId"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	CategorySlug string    `json:"categorySlug"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type UpsertEvent struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurredAt"`
	Data       Document  `json:"data"`
}
