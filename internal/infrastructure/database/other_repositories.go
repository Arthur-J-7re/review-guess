package database

import (
	"database/sql"
	"fmt"
	"time"

	"review-guess/internal/domain"
)

// ===== MOVIE SIMILARITY REPOSITORY =====

// MovieSimilarityRepository implémente domain.MovieSimilarityRepository
type MovieSimilarityRepositoryImpl struct {
	db *sql.DB
}

// NewMovieSimilarityRepository crée une nouvelle instance
func NewMovieSimilarityRepository(db *sql.DB) *MovieSimilarityRepositoryImpl {
	return &MovieSimilarityRepositoryImpl{db: db}
}

// Get récupère la similarité entre deux films
func (r *MovieSimilarityRepositoryImpl) Get(movieAID, movieBID string) (*domain.MovieSimilarity, error) {
	query := `
		SELECT movie_a, movie_b, similarity_score, shared_directors, shared_actors, shared_genres, year_proximity, calculated_at
		FROM movie_similarities WHERE (movie_a = ? AND movie_b = ?) OR (movie_a = ? AND movie_b = ?)
		LIMIT 1
	`

	var sim domain.MovieSimilarity
	err := r.db.QueryRow(query, movieAID, movieBID, movieBID, movieAID).Scan(
		&sim.MovieAID, &sim.MovieBID, &sim.SimilarityScore, &sim.SharedDirectors,
		&sim.SharedActors, &sim.SharedGenres, &sim.YearProximity, &sim.CalculatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get similarity: %w", err)
	}

	return &sim, nil
}

// GetTopSimilarMovies récupère les N films les plus similaires à un film donné
func (r *MovieSimilarityRepositoryImpl) GetTopSimilarMovies(movieID string, limit int) ([]*domain.MovieSimilarity, error) {
	query := `
		SELECT movie_a, movie_b, similarity_score, shared_directors, shared_actors, shared_genres, year_proximity, calculated_at
		FROM movie_similarities WHERE movie_a = ? OR movie_b = ?
		ORDER BY similarity_score DESC LIMIT ?
	`

	rows, err := r.db.Query(query, movieID, movieID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query similarities: %w", err)
	}
	defer rows.Close()

	var similarities []*domain.MovieSimilarity
	for rows.Next() {
		var sim domain.MovieSimilarity
		err := rows.Scan(
			&sim.MovieAID, &sim.MovieBID, &sim.SimilarityScore, &sim.SharedDirectors,
			&sim.SharedActors, &sim.SharedGenres, &sim.YearProximity, &sim.CalculatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan similarity: %w", err)
		}
		similarities = append(similarities, &sim)
	}

	return similarities, rows.Err()
}

// Create crée une nouvelle relation de similarité
func (r *MovieSimilarityRepositoryImpl) Create(similarity *domain.MovieSimilarity) error {
	query := `
		INSERT INTO movie_similarities (movie_a, movie_b, similarity_score, shared_directors, shared_actors, shared_genres, year_proximity, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	_, err := r.db.Exec(
		query,
		similarity.MovieAID, similarity.MovieBID, similarity.SimilarityScore,
		similarity.SharedDirectors, similarity.SharedActors, similarity.SharedGenres,
		similarity.YearProximity, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create similarity: %w", err)
	}

	similarity.CalculatedAt = now
	return nil
}

// CreateBatch crée plusieurs relations en une seule transaction
func (r *MovieSimilarityRepositoryImpl) CreateBatch(similarities []*domain.MovieSimilarity) error {
	if len(similarities) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO movie_similarities (movie_a, movie_b, similarity_score, shared_directors, shared_actors, shared_genres, year_proximity, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, sim := range similarities {
		_, err := stmt.Exec(
			sim.MovieAID, sim.MovieBID, sim.SimilarityScore,
			sim.SharedDirectors, sim.SharedActors, sim.SharedGenres,
			sim.YearProximity, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert similarity: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteByMovie supprime tous les enregistrements de similarité pour un film
func (r *MovieSimilarityRepositoryImpl) DeleteByMovie(movieID string) error {
	query := `DELETE FROM movie_similarities WHERE movie_a = ? OR movie_b = ?`

	_, err := r.db.Exec(query, movieID, movieID)
	if err != nil {
		return fmt.Errorf("failed to delete similarities: %w", err)
	}

	return nil
}

// ===== USER MOVIE REPOSITORY =====

// ReviewerMovieRepository implémente domain.ReviewerMovieRepository
type ReviewerMovieRepositoryImpl struct {
	db *sql.DB
}

// NewReviewerMovieRepository crée une nouvelle instance
func NewReviewerMovieRepository(db *sql.DB) *ReviewerMovieRepositoryImpl {
	return &ReviewerMovieRepositoryImpl{db: db}
}

// Get récupère l'enregistrement d'un film regardé par un reviewer
func (r *ReviewerMovieRepositoryImpl) Get(reviewerID, movieID string) (*domain.ReviewerMovie, error) {
	query := `
		SELECT reviewer_id, movie_id, has_review, rating, watched_at, created_at
		FROM reviewer_movies WHERE reviewer_id = ? AND movie_id = ?
	`

	var rm domain.ReviewerMovie
	err := r.db.QueryRow(query, reviewerID, movieID).Scan(
		&rm.ReviewerID, &rm.MovieID, &rm.HasReview, &rm.Rating, &rm.WatchedAt, &rm.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer movie: %w", err)
	}

	return &rm, nil
}

// GetMoviesWatchedByReviewer récupère tous les films regardés par un reviewer
func (r *ReviewerMovieRepositoryImpl) GetMoviesWatchedByReviewer(reviewerID string) ([]*domain.ReviewerMovie, error) {
	query := `
		SELECT reviewer_id, movie_id, has_review, rating, watched_at, created_at
		FROM reviewer_movies WHERE reviewer_id = ? ORDER BY watched_at DESC
	`

	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewer movies: %w", err)
	}
	defer rows.Close()

	var reviewerMovies []*domain.ReviewerMovie
	for rows.Next() {
		var rm domain.ReviewerMovie
		err := rows.Scan(
			&rm.ReviewerID, &rm.MovieID, &rm.HasReview, &rm.Rating, &rm.WatchedAt, &rm.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reviewer movie: %w", err)
		}
		reviewerMovies = append(reviewerMovies, &rm)
	}

	return reviewerMovies, rows.Err()
}

// GetMoviesNotWatchedByReviewer récupère les films NON regardés par un reviewer
func (r *ReviewerMovieRepositoryImpl) GetMoviesNotWatchedByReviewer(reviewerID string, limit int) ([]*domain.Movie, error) {
	query := `
		SELECT m.id, m.title, m.year, m.description, m.poster_url, m.tmdb_id, m.letterboxd_slug, m.created_at, m.updated_at
		FROM movies m
		WHERE m.id NOT IN (SELECT movie_id FROM reviewer_movies WHERE reviewer_id = ?)
		ORDER BY RANDOM() LIMIT ?
	`

	rows, err := r.db.Query(query, reviewerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unwatched movies: %w", err)
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

// Create crée un nouvel enregistrement
func (r *ReviewerMovieRepositoryImpl) Create(reviewerMovie *domain.ReviewerMovie) error {
	query := `
		INSERT INTO reviewer_movies (reviewer_id, movie_id, has_review, rating, watched_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	reviewerMovie.CreatedAt = now

	_, err := r.db.Exec(
		query, reviewerMovie.ReviewerID, reviewerMovie.MovieID, reviewerMovie.HasReview,
		reviewerMovie.Rating, reviewerMovie.WatchedAt, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create reviewer movie: %w", err)
	}

	return nil
}

// Update met à jour un enregistrement
func (r *ReviewerMovieRepositoryImpl) Update(reviewerMovie *domain.ReviewerMovie) error {
	query := `
		UPDATE reviewer_movies SET has_review = ?, rating = ?, watched_at = ? WHERE reviewer_id = ? AND movie_id = ?
	`

	_, err := r.db.Exec(
		query, reviewerMovie.HasReview, reviewerMovie.Rating, reviewerMovie.WatchedAt,
		reviewerMovie.ReviewerID, reviewerMovie.MovieID,
	)

	if err != nil {
		return fmt.Errorf("failed to update reviewer movie: %w", err)
	}

	return nil
}

// CreateBatch crée plusieurs enregistrements
func (r *ReviewerMovieRepositoryImpl) CreateBatch(reviewerMovies []*domain.ReviewerMovie) error {
	if len(reviewerMovies) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO reviewer_movies (reviewer_id, movie_id, has_review, rating, watched_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, rm := range reviewerMovies {
		rm.CreatedAt = now

		_, err := stmt.Exec(
			rm.ReviewerID, rm.MovieID, rm.HasReview,
			rm.Rating, rm.WatchedAt, now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert reviewer movie: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
