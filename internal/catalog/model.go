package catalog

import "time"

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Product struct {
	ID               string         `json:"id"`
	StoreID          string         `json:"storeId"`
	StoreName        string         `json:"storeName,omitempty"`
	StoreSlug        string         `json:"storeSlug"`
	Category         Category       `json:"category"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Description      string         `json:"description"`
	Status           string         `json:"status"`
	ModerationStatus string         `json:"moderationStatus"`
	ModerationNote   string         `json:"moderationNote"`
	Variants         []Variant      `json:"variants"`
	Images           []ProductImage `json:"images"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
}

type ProductInput struct {
	CategoryID  string `json:"categoryId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type Variant struct {
	ID         string            `json:"id"`
	SKU        string            `json:"sku"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
	PriceMinor int64             `json:"priceMinor"`
	Currency   string            `json:"currency"`
	Stock      int               `json:"stock"`
	Active     bool              `json:"active"`
}

type VariantInput struct {
	SKU        string            `json:"sku"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
	PriceMinor int64             `json:"priceMinor"`
	Currency   string            `json:"currency"`
	Stock      int               `json:"stock"`
}

type ProductImage struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	AltText  string `json:"altText"`
	Position int    `json:"position"`
}

type ImageInput struct {
	URL      string `json:"url"`
	AltText  string `json:"altText"`
	Position int    `json:"position"`
}

type Filters struct {
	Search       string
	CategorySlug string
	StoreSlug    string
	MinPrice     *int64
	MaxPrice     *int64
	InStock      bool
	Page         int
	PageSize     int
}

type Summary struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Slug          string   `json:"slug"`
	StoreName     string   `json:"storeName"`
	StoreSlug     string   `json:"storeSlug"`
	Category      Category `json:"category"`
	MinPriceMinor int64    `json:"minPriceMinor"`
	Currency      string   `json:"currency"`
	InStock       bool     `json:"inStock"`
	ImageURL      string   `json:"imageUrl"`
}

type Page struct {
	Items    []Summary `json:"items"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Total    int       `json:"total"`
}
