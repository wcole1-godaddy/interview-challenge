package main

import "time"

// dbProduct is the internal representation matching the SQLite schema.
// Prices are stored as integer cents.
type dbProduct struct {
	ID          int
	Name        string
	Description string
	PriceCents  int
	Category    string
	InStock     bool
	Quantity    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Product is the API-facing representation.
// Prices are represented as dollar floats (e.g., 29.99).
type Product struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Category    string     `json:"category"`
	InStock     bool       `json:"in_stock"`
	Quantity    int        `json:"quantity"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// CreateProductRequest is the expected body for POST /products.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	InStock     bool    `json:"in_stock"`
	Quantity    int     `json:"quantity"`
}
