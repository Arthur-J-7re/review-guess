package database

import (
	"database/sql"
	"fmt"
	"time"

	"review-guess/internal/domain"
)

// ReviewRepository implémente domain.ReviewRepository
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository crée une nouvelle instance
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Get récupère une review par ID
func (r *ReviewRepository) Get(id string) (*domain.Review, error) {
	query := `
		SELECT id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at
		FROM reviews WHERE id = ?
	`

	var review domain.Review
	err := r.db.QueryRow(query, id).Scan(
		&review.ID, &review.ReviewerID, &review.MovieID, &review.Title, &review.Content,
		&review.Rating, &review.Liked, &review.Spoilers, &review.Usable,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	return &review, nil
}

// GetByReviewerAndMovie récupère la review d'un reviewer pour un film
func (r *ReviewRepository) GetByReviewerAndMovie(reviewerID, movieID string) (*domain.Review, error) {
	query := `
		SELECT id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at
		FROM reviews WHERE reviewer_id = ? AND movie_id = ?
	`

	var review domain.Review
	err := r.db.QueryRow(query, reviewerID, movieID).Scan(
		&review.ID, &review.ReviewerID, &review.MovieID, &review.Title, &review.Content,
		&review.Rating, &review.Liked, &review.Spoilers, &review.Usable,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get review by reviewer and movie: %w", err)
	}

	return &review, nil
}

// GetUsableReviewsForReviewer récupère toutes les reviews utilisables d'un reviewer
func (r *ReviewRepository) GetUsableReviewsForReviewer(reviewerID string) ([]*domain.Review, error) {
	query := `
		SELECT id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at
		FROM reviews WHERE reviewer_id = ? AND usable = TRUE ORDER BY created_at
	`

	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*domain.Review
	for rows.Next() {
		var review domain.Review
		err := rows.Scan(
			&review.ID, &review.ReviewerID, &review.MovieID, &review.Title, &review.Content,
			&review.Rating, &review.Liked, &review.Spoilers, &review.Usable,
			&review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan review: %w", err)
		}
		reviews = append(reviews, &review)
	}

	return reviews, rows.Err()
}

// GetRandomUsableReview récupère une review aléatoire marquée comme usable
func (r *ReviewRepository) GetRandomUsableReview() (*domain.Review, error) {
	query := `
		SELECT id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at
		FROM reviews WHERE usable = TRUE ORDER BY RANDOM() LIMIT 1
	`

	var review domain.Review
	err := r.db.QueryRow(query).Scan(
		&review.ID, &review.ReviewerID, &review.MovieID, &review.Title, &review.Content,
		&review.Rating, &review.Liked, &review.Spoilers, &review.Usable,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get random review: %w", err)
	}

	return &review, nil
}

// Create crée une nouvelle review
func (r *ReviewRepository) Create(review *domain.Review) error {
	query := `
		INSERT INTO reviews (id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		review.ID, review.ReviewerID, review.MovieID, review.Title, review.Content,
		review.Rating, review.Liked, review.Spoilers, review.Usable,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	return nil
}

// Update met à jour une review
func (r *ReviewRepository) Update(review *domain.Review) error {
	query := `
		UPDATE reviews
		SET title = ?, content = ?, rating = ?, liked = ?, spoilers = ?, usable = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	review.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		review.Title, review.Content, review.Rating, review.Liked, review.Spoilers, review.Usable,
		now, review.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update review: %w", err)
	}

	return nil
}

// CreateBatch crée plusieurs reviews en une seule transaction
func (r *ReviewRepository) CreateBatch(reviews []*domain.Review) error {
	if len(reviews) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO reviews (id, reviewer_id, movie_id, title, content, rating, liked, spoilers, usable, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, review := range reviews {
		review.CreatedAt = now
		review.UpdatedAt = now

		_, err := stmt.Exec(
			review.ID, review.ReviewerID, review.MovieID, review.Title, review.Content,
			review.Rating, review.Liked, review.Spoilers, review.Usable,
			now, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert review: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
