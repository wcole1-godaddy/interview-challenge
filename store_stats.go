package main

import (
	"fmt"
	"time"
)

// createAuditTable creates the audit_log table if it doesn't exist.
func createAuditTable(s *Store) error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			detail TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id)
		)
	`)
	return err
}

// LogAudit records an action taken on a product.
func (s *Store) LogAudit(productID int, action, detail string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(
		`INSERT INTO audit_log (product_id, action, detail, created_at) VALUES (?, ?, ?, ?)`,
		productID, action, detail, now,
	)
	return err
}

// GetAuditLog returns the audit trail for a specific product.
func (s *Store) GetAuditLog(productID int) ([]AuditEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, product_id, action, detail, created_at
		 FROM audit_log WHERE product_id = ? ORDER BY created_at DESC`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("get audit log: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		err := rows.Scan(&e.ID, &e.ProductID, &e.Action, &e.Detail, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetRecentAuditLog returns the most recent audit entries across all products.
func (s *Store) GetRecentAuditLog(limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(
		`SELECT id, product_id, action, detail, created_at
		 FROM audit_log ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent audit log: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		err := rows.Scan(&e.ID, &e.ProductID, &e.Action, &e.Detail, &e.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan audit entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetCategoryStats returns aggregate statistics grouped by category.
func (s *Store) GetCategoryStats() ([]CategoryStat, error) {
	rows, err := s.db.Query(`
		SELECT category, COUNT(*) as product_count,
		       COALESCE(AVG(price_cents), 0) as avg_price,
		       COALESCE(SUM(quantity), 0) as total_inventory,
		       SUM(CASE WHEN in_stock = 1 THEN 1 ELSE 0 END) as in_stock_count
		FROM products
		WHERE deleted_at IS NULL
		GROUP BY category
		ORDER BY product_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("category stats: %w", err)
	}
	defer rows.Close()

	var stats []CategoryStat
	for rows.Next() {
		var s CategoryStat
		err := rows.Scan(&s.Category, &s.ProductCount, &s.AveragePrice, &s.TotalInventory, &s.InStockCount)
		if err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// GetProductCount returns counts of total, in-stock, and out-of-stock products.
func (s *Store) GetProductCount() (total, inStock, outOfStock int, err error) {
	err = s.db.QueryRow(`
		SELECT COUNT(*),
		       SUM(CASE WHEN in_stock = 1 THEN 1 ELSE 0 END),
		       SUM(CASE WHEN in_stock = 0 THEN 1 ELSE 0 END)
		FROM products WHERE deleted_at IS NULL
	`).Scan(&total, &inStock, &outOfStock)
	return
}

// GetAverageProductPrice returns the average price in cents across all active products.
func (s *Store) GetAverageProductPrice() (float64, error) {
	var avg float64
	err := s.db.QueryRow(
		`SELECT COALESCE(AVG(price_cents), 0) FROM products WHERE deleted_at IS NULL`,
	).Scan(&avg)
	return avg, err
}

// GetTotalInventory returns the sum of all product quantities.
func (s *Store) GetTotalInventory() (int, error) {
	var total int
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM products WHERE deleted_at IS NULL`,
	).Scan(&total)
	return total, err
}

// GetTotalReviewCount returns the total number of reviews in the system.
func (s *Store) GetTotalReviewCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&count)
	return count, err
}

// SearchProducts performs a basic text search across product name and description.
func (s *Store) SearchProducts(query string) ([]dbProduct, error) {
	pattern := "%" + query + "%"
	rows, err := s.db.Query(
		`SELECT id, name, description, price_cents, category, in_stock, quantity, created_at, updated_at, deleted_at
		 FROM products
		 WHERE deleted_at IS NULL AND (name LIKE ? OR description LIKE ?)
		 ORDER BY name`,
		pattern, pattern,
	)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
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

// ListCategories returns distinct category names.
func (s *Store) ListCategories() ([]string, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT category FROM products WHERE deleted_at IS NULL AND category != '' ORDER BY category`,
	)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}
