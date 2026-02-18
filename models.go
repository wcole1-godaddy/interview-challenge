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

// dbReview is the internal representation for product reviews.
type dbReview struct {
	ID        int
	ProductID int
	Author    string
	Rating    int
	Comment   string
	Approved  bool
	CreatedAt time.Time
}

// Review is the API-facing representation of a product review.
type Review struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Author    string    `json:"author"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	Approved  bool      `json:"approved"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateReviewRequest is the expected body for POST /products/:id/reviews.
type CreateReviewRequest struct {
	Author  string `json:"author"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// ProductWithReviews combines a product with its reviews for detail views.
type ProductWithReviews struct {
	Product       Product  `json:"product"`
	Reviews       []Review `json:"reviews"`
	AverageRating float64  `json:"average_rating"`
	ReviewCount   int      `json:"review_count"`
}

// CategoryStat holds aggregate statistics for a product category.
type CategoryStat struct {
	Category       string  `json:"category"`
	ProductCount   int     `json:"product_count"`
	AveragePrice   float64 `json:"average_price"`
	TotalInventory int     `json:"total_inventory"`
	InStockCount   int     `json:"in_stock_count"`
}

// DashboardStats holds overall catalog statistics.
type DashboardStats struct {
	TotalProducts   int            `json:"total_products"`
	TotalInStock    int            `json:"total_in_stock"`
	TotalOutOfStock int            `json:"total_out_of_stock"`
	AveragePrice    float64        `json:"average_price"`
	TotalInventory  int            `json:"total_inventory"`
	TotalReviews    int            `json:"total_reviews"`
	Categories      []CategoryStat `json:"categories"`
}

// AuditEntry represents a logged change to a product.
type AuditEntry struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// HealthStatus is returned by the health check endpoint.
type HealthStatus struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
}

// dbVariant is the internal representation for product variants.
// Each variant represents a specific combination of attributes (e.g., size + color).
// Prices are stored as integer cents; a zero price means "use the parent product price".
type dbVariant struct {
	ID         int
	ProductID  int
	SKU        string
	Name       string
	PriceCents int
	Quantity   int
	InStock    bool
	Attributes string // JSON-encoded key-value pairs, e.g., {"size":"L","color":"blue"}
	SortOrder  int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Variant is the API-facing representation of a product variant.
type Variant struct {
	ID         int                    `json:"id"`
	ProductID  int                    `json:"product_id"`
	SKU        string                 `json:"sku"`
	Name       string                 `json:"name"`
	Price      float64                `json:"price"`
	Quantity   int                    `json:"quantity"`
	InStock    bool                   `json:"in_stock"`
	Attributes map[string]string      `json:"attributes"`
	SortOrder  int                    `json:"sort_order"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// CreateVariantRequest is the expected body for POST /products/:id/variants.
type CreateVariantRequest struct {
	SKU        string            `json:"sku"`
	Name       string            `json:"name"`
	Price      float64           `json:"price"`
	Quantity   int               `json:"quantity"`
	Attributes map[string]string `json:"attributes"`
	SortOrder  int               `json:"sort_order"`
}

// UpdateVariantRequest is the expected body for PUT /products/:id/variants/:variantId.
type UpdateVariantRequest struct {
	SKU        string            `json:"sku"`
	Name       string            `json:"name"`
	Price      float64           `json:"price"`
	Quantity   int               `json:"quantity"`
	InStock    bool              `json:"in_stock"`
	Attributes map[string]string `json:"attributes"`
	SortOrder  int               `json:"sort_order"`
}

// ProductDetail is the full product view including variants and reviews.
type ProductDetail struct {
	Product       Product   `json:"product"`
	Variants      []Variant `json:"variants"`
	Reviews       []Review  `json:"reviews"`
	AverageRating float64   `json:"average_rating"`
	ReviewCount   int       `json:"review_count"`
	TotalStock    int       `json:"total_stock"`
}

// VariantInventory summarizes stock across all variants for a product.
type VariantInventory struct {
	ProductID    int `json:"product_id"`
	VariantCount int `json:"variant_count"`
	TotalStock   int `json:"total_stock"`
	InStockCount int `json:"in_stock_count"`
}
