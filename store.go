package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("create tables: %w", err)
	}

	return &Store{
		db:           db,
		productCache: make(map[int]cachedProduct),
	}, nil
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
	now := time.Now().UTC()
	_, err := s.db.Exec(
		`UPDATE products SET quantity = quantity - 1, in_stock = (quantity - 1 > 0), updated_at = ? WHERE id = ?`,
		now, id,
	)
	return err
}
