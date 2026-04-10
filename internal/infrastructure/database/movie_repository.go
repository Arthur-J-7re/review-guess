package database

import (
	"database/sql"
	"fmt"
	"time"

	"review-guess/internal/domain"
)

// MovieRepository implémente domain.MovieRepository
type MovieRepository struct {
	db *sql.DB
}

// NewMovieRepository crée une nouvelle instance
func NewMovieRepository(db *sql.DB) *MovieRepository {
	return &MovieRepository{db: db}
}

// Get récupère un film par ID
func (r *MovieRepository) Get(id string) (*domain.Movie, error) {
	query := `
		SELECT id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at
		FROM movies WHERE id = ?
	`

	var movie domain.Movie
	err := r.db.QueryRow(query, id).Scan(
		&movie.ID, &movie.Title, &movie.Year, &movie.Description,
		&movie.PosterURL, &movie.TmdbID, &movie.LetterboxdSlug,
		&movie.CreatedAt, &movie.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	return &movie, nil
}

// GetByTmdbID récupère un film par son ID TMDB
func (r *MovieRepository) GetByTmdbID(tmdbID int) (*domain.Movie, error) {
	query := `
		SELECT id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at
		FROM movies WHERE tmdb_id = ?
	`

	var movie domain.Movie
	err := r.db.QueryRow(query, tmdbID).Scan(
		&movie.ID, &movie.Title, &movie.Year, &movie.Description,
		&movie.PosterURL, &movie.TmdbID, &movie.LetterboxdSlug,
		&movie.CreatedAt, &movie.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get movie by TMDB ID: %w", err)
	}

	return &movie, nil
}

// GetByLetterboxdSlug récupère un film par son slug Letterboxd
func (r *MovieRepository) GetByLetterboxdSlug(slug string) (*domain.Movie, error) {
	query := `
		SELECT id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at
		FROM movies WHERE letterboxd_slug = ?
	`

	var movie domain.Movie
	err := r.db.QueryRow(query, slug).Scan(
		&movie.ID, &movie.Title, &movie.Year, &movie.Description,
		&movie.PosterURL, &movie.TmdbID, &movie.LetterboxdSlug,
		&movie.CreatedAt, &movie.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get movie by slug: %w", err)
	}

	return &movie, nil
}

// Create crée un nouveau film
func (r *MovieRepository) Create(movie *domain.Movie) error {
	query := `
		INSERT INTO movies (id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.Exec(
		query,
		movie.ID, movie.Title, movie.Year, movie.Description,
		movie.PosterURL, movie.TmdbID, movie.LetterboxdSlug,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create movie: %w", err)
	}

	movie.CreatedAt = now
	movie.UpdatedAt = now
	return nil
}

// Update met à jour un film
func (r *MovieRepository) Update(movie *domain.Movie) error {
	query := `
		UPDATE movies
		SET title = ?, year = ?, description = ?, poster_url = ?, tmdb_id = ?, letterboxd_slug = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := r.db.Exec(
		query,
		movie.Title, movie.Year, movie.Description,
		movie.PosterURL, movie.TmdbID, movie.LetterboxdSlug,
		now, movie.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update movie: %w", err)
	}

	movie.UpdatedAt = now
	return nil
}

// GetAll récupère tous les films
func (r *MovieRepository) GetAll() ([]*domain.Movie, error) {
	query := `
		SELECT id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at
		FROM movies ORDER BY title
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

	var movies []*domain.Movie
	for rows.Next() {
		var movie domain.Movie
		err := rows.Scan(
			&movie.ID, &movie.Title, &movie.Year, &movie.Description,
			&movie.PosterURL, &movie.TmdbID, &movie.LetterboxdSlug,
			&movie.CreatedAt, &movie.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}
		movies = append(movies, &movie)
	}

	return movies, rows.Err()
}

// GetMany récupère plusieurs films par leurs IDs
func (r *MovieRepository) GetMany(ids []string) ([]*domain.Movie, error) {
	if len(ids) == 0 {
		return []*domain.Movie{}, nil
	}

	// Build IN clause dynamically
	query := `
		SELECT id, title, year, description, poster_url, tmdb_id, letterboxd_slug, created_at, updated_at
		FROM movies WHERE id IN (`

	var args []interface{}
	for i, id := range ids {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args = append(args, id)
	}
	query += `) ORDER BY title`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

	var movies []*domain.Movie
	for rows.Next() {
		var movie domain.Movie
		err := rows.Scan(
			&movie.ID, &movie.Title, &movie.Year, &movie.Description,
			&movie.PosterURL, &movie.TmdbID, &movie.LetterboxdSlug,
			&movie.CreatedAt, &movie.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}
		movies = append(movies, &movie)
	}

	return movies, rows.Err()
}
