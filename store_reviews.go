package main

import (
	"fmt"
	"time"
)

// CreateReviewTable creates the reviews table if it doesn't exist.
func createReviewTable(s *Store) error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			author TEXT NOT NULL,
			rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
			comment TEXT DEFAULT '',
			approved BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id)
		)
	`)
	return err
}

// CreateReview inserts a new review for a product.
func (s *Store) CreateReview(productID int, author string, rating int, comment string) (int, error) {
	if author == "" {
		return 0, fmt.Errorf("author is required")
	}
	if rating < 1 || rating > 5 {
		return 0, fmt.Errorf("rating must be between 1 and 5")
	}

	// Verify product exists and is not deleted.
	_, err := s.GetProduct(productID)
	if err != nil {
		return 0, fmt.Errorf("product not found")
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(
		`INSERT INTO reviews (product_id, author, rating, comment, approved, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		productID, author, rating, comment, false, now,
	)
	if err != nil {
		return 0, fmt.Errorf("insert review: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// ListReviews returns all reviews for a product.
func (s *Store) ListReviews(productID int) ([]dbReview, error) {
	rows, err := s.db.Query(
		`SELECT id, product_id, author, rating, comment, approved, created_at
		 FROM reviews WHERE product_id = ? ORDER BY created_at DESC`,
		productID,
	)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []dbReview
	for rows.Next() {
		var r dbReview
		err := rows.Scan(&r.ID, &r.ProductID, &r.Author, &r.Rating, &r.Comment, &r.Approved, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

// GetReview returns a single review by ID.
func (s *Store) GetReview(reviewID int) (*dbReview, error) {
	var r dbReview
	err := s.db.QueryRow(
		`SELECT id, product_id, author, rating, comment, approved, created_at
		 FROM reviews WHERE id = ?`,
		reviewID,
	).Scan(&r.ID, &r.ProductID, &r.Author, &r.Rating, &r.Comment, &r.Approved, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteReview removes a review by ID.
func (s *Store) DeleteReview(reviewID int) error {
	result, err := s.db.Exec(`DELETE FROM reviews WHERE id = ?`, reviewID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("review not found")
	}
	return nil
}

// ApproveReview marks a review as approved.
func (s *Store) ApproveReview(reviewID int) error {
	result, err := s.db.Exec(`UPDATE reviews SET approved = 1 WHERE id = ?`, reviewID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("review not found")
	}
	return nil
}

// GetAverageRating returns the average rating for a product.
func (s *Store) GetAverageRating(productID int) (float64, int, error) {
	var avg float64
	var count int
	err := s.db.QueryRow(
		`SELECT COALESCE(AVG(rating), 0), COUNT(*) FROM reviews WHERE product_id = ?`,
		productID,
	).Scan(&avg, &count)
	if err != nil {
		return 0, 0, err
	}
	return avg, count, nil
}

// GetRecentReviews returns the most recent reviews across all products.
func (s *Store) GetRecentReviews(limit int) ([]dbReview, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(
		`SELECT id, product_id, author, rating, comment, approved, created_at
		 FROM reviews ORDER BY created_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent reviews: %w", err)
	}
	defer rows.Close()

	var reviews []dbReview
	for rows.Next() {
		var r dbReview
		err := rows.Scan(&r.ID, &r.ProductID, &r.Author, &r.Rating, &r.Comment, &r.Approved, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

// CountReviewsByProduct returns review counts grouped by product.
func (s *Store) CountReviewsByProduct() (map[int]int, error) {
	rows, err := s.db.Query(
		`SELECT product_id, COUNT(*) FROM reviews GROUP BY product_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int]int)
	for rows.Next() {
		var productID, count int
		if err := rows.Scan(&productID, &count); err != nil {
			return nil, err
		}
		counts[productID] = count
	}
	return counts, rows.Err()
}
