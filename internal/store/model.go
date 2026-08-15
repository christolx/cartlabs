package store

import "time"

type Store struct {
	ID             string    `json:"id"`
	SellerID       string    `json:"sellerId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	ModerationNote string    `json:"moderationNote"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Input struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type Profile struct {
	ID                string
	Name              string
	Slug              string
	Description       string
	SellerDisplayName string
	CreatedAt         time.Time
}
