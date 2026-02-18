package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db           *sql.DB
	cacheMu      sync.RWMutex
	productCache map[int]cachedProduct
}

type cachedProduct struct {
	product   dbProduct
	expiresAt time.Time
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	store := &Store{
		db:           db,
		productCache: make(map[int]cachedProduct),
	}

	if err := createReviewTable(store); err != nil {
		return nil, fmt.Errorf("create review table: %w", err)
	}

	if err := createAuditTable(store); err != nil {
		return nil, fmt.Errorf("create audit table: %w", err)
	}

	if err := createVariantTable(store); err != nil {
		return nil, fmt.Errorf("create variant table: %w", err)
	}

	if err := seedData(db); err != nil {
		return nil, fmt.Errorf("seed data: %w", err)
	}

	return store, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT DEFAULT '',
			price_cents INTEGER NOT NULL,
			category TEXT DEFAULT '',
			in_stock BOOLEAN DEFAULT 1,
			quantity INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL,
			UNIQUE(name)
		)
	`)
	return err
}

func seedData(db *sql.DB) error {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	log.Println("Seeding database with sample products...")

	now := time.Now().UTC()
	seeds := []struct {
		name, desc    string
		priceCents    int
		category      string
		inStock       bool
		quantity      int
	}{
		{"Wireless Mouse", "Ergonomic wireless mouse with USB receiver", 2499, "electronics", true, 25},
		{"Mechanical Keyboard", "Cherry MX Blue switches, full-size layout", 8999, "electronics", true, 12},
		{"USB-C Hub", "7-in-1 USB-C adapter with HDMI and ethernet", 3499, "electronics", true, 0},
		{"Standing Desk", "Electric sit-stand desk, 60 inch wide", 49999, "furniture", false, 8},
		{"Monitor Arm", "Single monitor mount, gas spring, VESA compatible", 4999, "furniture", true, 0},
		{"Notebook Pack", "200-page lined notebooks, pack of 3", -500, "office", true, 50},
		{"Desk Lamp", "LED desk lamp with adjustable brightness", 3299, "office", true, 15},
		{"Webcam HD", "1080p webcam with built-in microphone", 5999, "electronics", true, 3},
	}

	for _, s := range seeds {
		_, err := db.Exec(
			`INSERT OR IGNORE INTO products (name, description, price_cents, category, in_stock, quantity, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			s.name, s.desc, s.priceCents, s.category, s.inStock, s.quantity, now, now,
		)
		if err != nil {
			return err
		}
	}

	// Seed variants for select products.
	var variantCount int
	db.QueryRow(`SELECT COUNT(*) FROM variants`).Scan(&variantCount)
	if variantCount == 0 {
		variantSeeds := []struct {
			productID  int
			sku, name  string
			priceCents int
			quantity   int
			attrs      string
			sortOrder  int
		}{
			{1, "WM-BLK", "Wireless Mouse - Black", 2499, 10, `{"color":"black"}`, 1},
			{1, "WM-WHT", "Wireless Mouse - White", 2499, 8, `{"color":"white"}`, 2},
			{1, "WM-BLU", "Wireless Mouse - Blue", 2699, 7, `{"color":"blue"}`, 3},
			{2, "KB-FULL", "Mechanical Keyboard - Full Size", 8999, 6, `{"size":"full"}`, 1},
			{2, "KB-TKL", "Mechanical Keyboard - Tenkeyless", 7999, 4, `{"size":"tenkeyless"}`, 2},
			{2, "KB-65", "Mechanical Keyboard - 65%", 8499, 2, `{"size":"65%"}`, 3},
			{4, "SD-48", "Standing Desk - 48 inch", 39999, 3, `{"width":"48in"}`, 1},
			{4, "SD-60", "Standing Desk - 60 inch", 49999, 5, `{"width":"60in"}`, 2},
			{4, "SD-72", "Standing Desk - 72 inch", 59999, 0, `{"width":"72in"}`, 3},
		}

		for _, v := range variantSeeds {
			inStock := v.quantity > 0
			_, err := db.Exec(
				`INSERT OR IGNORE INTO variants (product_id, sku, name, price_cents, quantity, in_stock, attributes, sort_order, created_at, updated_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				v.productID, v.sku, v.name, v.priceCents, v.quantity, inStock, v.attrs, v.sortOrder, now, now,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// ListProducts returns all products, optionally filtered by category.
func (s *Store) ListProducts(category string) ([]dbProduct, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = fmt.Sprintf(`SELECT id, name, description, price_cents, category, in_stock, quantity, created_at, updated_at, deleted_at FROM products WHERE category = '%s'`, category)
	} else {
		query = `SELECT id, name, description, price_cents, category, in_stock, quantity, created_at, updated_at, deleted_at FROM products`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var products []dbProduct
	for rows.Next() {
		var p dbProduct
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Category,
			&p.InStock, &p.Quantity, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
		if err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}

	return products, rows.Err()
}

func (s *Store) GetProduct(id int) (*dbProduct, error) {
	now := time.Now().UTC()
	s.cacheMu.RLock()
	entry, ok := s.productCache[id]
	s.cacheMu.RUnlock()
	if ok && now.Before(entry.expiresAt) {
		p := entry.product
		return &p, nil
	}

	var p dbProduct
	err := s.db.QueryRow(
		`SELECT id, name, description, price_cents, category, in_stock, quantity, created_at, updated_at, deleted_at
		 FROM products WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.Category,
		&p.InStock, &p.Quantity, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}

	s.cacheMu.Lock()
	s.productCache[id] = cachedProduct{
		product:   p,
		expiresAt: now.Add(3 * time.Second),
	}
	s.cacheMu.Unlock()
	return &p, nil
}

func (s *Store) CreateProduct(name, description string, priceCents int, category string, inStock bool, quantity int) (int, error) {
	if name == "" {
		return 0, fmt.Errorf("name is required")
	}
	if priceCents < 0 {
		return 0, fmt.Errorf("price must be non-negative")
	}

	if len(description) > 128 {
		log.Printf("WARN: description for %q truncated from %d to 128 characters", name, len(description))
		description = description[:128]
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO products (name, description, price_cents, category, in_stock, quantity, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		name, description, priceCents, category, inStock, quantity, now, now,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateProduct updates fields for a product.
func (s *Store) UpdateProduct(id int, name, description string, priceCents int, category string, inStock bool, quantity int) error {
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`UPDATE products SET name = ?, description = ?, price_cents = ?, category = ?, in_stock = ?, quantity = ?, updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL`,
		name, description, priceCents, category, inStock, quantity, now, id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (s *Store) DeleteProduct(id int) error {
	now := time.Now().UTC()
	result, err := s.db.Exec(
		`UPDATE products SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`,
		now, now, id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// DecrementQuantity decreases quantity by 1 and updates in_stock.
func (s *Store) DecrementQuantity(id int) error {
	var currentQty int
	err := s.db.QueryRow(`SELECT quantity FROM products WHERE id = ?`, id).Scan(&currentQty)
	if err != nil {
		return err
	}

	newQty := currentQty - 1
	inStock := newQty > 0
	now := time.Now().UTC()

	_, err = s.db.Exec(
		`UPDATE products SET quantity = ?, in_stock = ?, updated_at = ? WHERE id = ?`,
		newQty, inStock, now, id,
	)
	return err
}
