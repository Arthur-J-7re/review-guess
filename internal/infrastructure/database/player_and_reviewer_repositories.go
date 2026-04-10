package database

import (
	"database/sql"
	"fmt"
	"time"

	"review-guess/internal/domain"
)

// PlayerRepository implémente domain.PlayerRepository
type PlayerRepositoryImpl struct {
	db *sql.DB
}

// NewPlayerRepository crée une nouvelle instance
func NewPlayerRepository(db *sql.DB) *PlayerRepositoryImpl {
	return &PlayerRepositoryImpl{db: db}
}

// Get récupère un joueur par ID
func (r *PlayerRepositoryImpl) Get(id string) (*domain.Player, error) {
	query := `
		SELECT id, nickname, is_logged_in, last_reviewed_reviewer_id, created_at, last_played_at, updated_at
		FROM players WHERE id = ?
	`

	var player domain.Player
	var nickname sql.NullString
	var lastReviewedReviewerID sql.NullString
	var lastPlayedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&player.ID, &nickname, &player.IsLoggedIn, &lastReviewedReviewerID,
		&player.CreatedAt, &lastPlayedAt, &player.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	if nickname.Valid {
		player.Nickname = nickname.String
	}
	if lastReviewedReviewerID.Valid {
		player.LastReviewedReviewerID = &lastReviewedReviewerID.String
	}
	if lastPlayedAt.Valid {
		player.LastPlayedAt = &lastPlayedAt.Time
	}

	return &player, nil
}

// Create crée un nouveau joueur
func (r *PlayerRepositoryImpl) Create(player *domain.Player) error {
	query := `
		INSERT INTO players (id, nickname, is_logged_in, last_reviewed_reviewer_id, created_at, last_played_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	player.CreatedAt = now
	player.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		player.ID, player.Nickname, player.IsLoggedIn, player.LastReviewedReviewerID,
		now, player.LastPlayedAt, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create player: %w", err)
	}

	return nil
}

// Update met à jour un joueur
func (r *PlayerRepositoryImpl) Update(player *domain.Player) error {
	query := `
		UPDATE players
		SET nickname = ?, is_logged_in = ?, last_reviewed_reviewer_id = ?, last_played_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	player.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		player.Nickname, player.IsLoggedIn, player.LastReviewedReviewerID, player.LastPlayedAt, now,
		player.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
	}

	return nil
}

// List récupère tous les joueurs
func (r *PlayerRepositoryImpl) List() ([]*domain.Player, error) {
	query := `
		SELECT id, nickname, is_logged_in, last_reviewed_reviewer_id, created_at, last_played_at, updated_at
		FROM players ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []*domain.Player
	for rows.Next() {
		var player domain.Player
		var nickname sql.NullString
		var lastReviewedReviewerID sql.NullString
		var lastPlayedAt sql.NullTime

		err := rows.Scan(
			&player.ID, &nickname, &player.IsLoggedIn, &lastReviewedReviewerID,
			&player.CreatedAt, &lastPlayedAt, &player.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}

		if nickname.Valid {
			player.Nickname = nickname.String
		}
		if lastReviewedReviewerID.Valid {
			player.LastReviewedReviewerID = &lastReviewedReviewerID.String
		}
		if lastPlayedAt.Valid {
			player.LastPlayedAt = &lastPlayedAt.Time
		}

		players = append(players, &player)
	}

	return players, rows.Err()
}

// ===== LETTERBOXD REVIEWER REPOSITORY =====

// LetterboxdReviewerRepositoryImpl implémente domain.LetterboxdReviewerRepository
type LetterboxdReviewerRepositoryImpl struct {
	db *sql.DB
}

// NewLetterboxdReviewerRepository crée une nouvelle instance
func NewLetterboxdReviewerRepository(db *sql.DB) *LetterboxdReviewerRepositoryImpl {
	return &LetterboxdReviewerRepositoryImpl{db: db}
}

// Get récupère un reviewer par ID
func (r *LetterboxdReviewerRepositoryImpl) Get(id string) (*domain.LetterboxdReviewer, error) {
	query := `
		SELECT id, letterboxd_username, total_reviews, total_movies_watched, 
		       last_review_page_scrapped, last_movie_page_scrapped, last_scrapped_at, created_at, updated_at
		FROM letterboxd_reviewers WHERE id = ?
	`

	var reviewer domain.LetterboxdReviewer
	var lastScrappedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&reviewer.ID, &reviewer.LetterboxdUsername, &reviewer.TotalReviews, &reviewer.TotalMoviesWatched,
		&reviewer.LastReviewPageScrapped, &reviewer.LastMoviePageScrapped, &lastScrappedAt,
		&reviewer.CreatedAt, &reviewer.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}

	if lastScrappedAt.Valid {
		reviewer.LastScrappedAt = &lastScrappedAt.Time
	}

	return &reviewer, nil
}

// GetByUsername récupère un reviewer par son username Letterboxd
func (r *LetterboxdReviewerRepositoryImpl) GetByUsername(username string) (*domain.LetterboxdReviewer, error) {
	query := `
		SELECT id, letterboxd_username, total_reviews, total_movies_watched, 
		       last_review_page_scrapped, last_movie_page_scrapped, last_scrapped_at, created_at, updated_at
		FROM letterboxd_reviewers WHERE letterboxd_username = ?
	`

	var reviewer domain.LetterboxdReviewer
	var lastScrappedAt sql.NullTime

	err := r.db.QueryRow(query, username).Scan(
		&reviewer.ID, &reviewer.LetterboxdUsername, &reviewer.TotalReviews, &reviewer.TotalMoviesWatched,
		&reviewer.LastReviewPageScrapped, &reviewer.LastMoviePageScrapped, &lastScrappedAt,
		&reviewer.CreatedAt, &reviewer.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get reviewer: %w", err)
	}

	if lastScrappedAt.Valid {
		reviewer.LastScrappedAt = &lastScrappedAt.Time
	}

	return &reviewer, nil
}

// Create crée un nouveau reviewer
func (r *LetterboxdReviewerRepositoryImpl) Create(reviewer *domain.LetterboxdReviewer) error {
	query := `
		INSERT INTO letterboxd_reviewers (id, letterboxd_username, total_reviews, total_movies_watched, 
		                                 last_review_page_scrapped, last_movie_page_scrapped, last_scrapped_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	reviewer.CreatedAt = now
	reviewer.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		reviewer.ID, reviewer.LetterboxdUsername, reviewer.TotalReviews, reviewer.TotalMoviesWatched,
		reviewer.LastReviewPageScrapped, reviewer.LastMoviePageScrapped, reviewer.LastScrappedAt,
		now, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create reviewer: %w", err)
	}

	return nil
}

// Update met à jour un reviewer
func (r *LetterboxdReviewerRepositoryImpl) Update(reviewer *domain.LetterboxdReviewer) error {
	query := `
		UPDATE letterboxd_reviewers
		SET total_reviews = ?, total_movies_watched = ?, 
		    last_review_page_scrapped = ?, last_movie_page_scrapped = ?, 
		    last_scrapped_at = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	reviewer.UpdatedAt = now

	_, err := r.db.Exec(
		query,
		reviewer.TotalReviews, reviewer.TotalMoviesWatched,
		reviewer.LastReviewPageScrapped, reviewer.LastMoviePageScrapped,
		reviewer.LastScrappedAt, now, reviewer.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update reviewer: %w", err)
	}

	return nil
}

// List récupère tous les reviewers
func (r *LetterboxdReviewerRepositoryImpl) List() ([]*domain.LetterboxdReviewer, error) {
	query := `
		SELECT id, letterboxd_username, total_reviews, total_movies_watched, 
		       last_review_page_scrapped, last_movie_page_scrapped, last_scrapped_at, created_at, updated_at
		FROM letterboxd_reviewers ORDER BY letterboxd_username
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []*domain.LetterboxdReviewer
	for rows.Next() {
		var reviewer domain.LetterboxdReviewer
		var lastScrappedAt sql.NullTime

		err := rows.Scan(
			&reviewer.ID, &reviewer.LetterboxdUsername, &reviewer.TotalReviews, &reviewer.TotalMoviesWatched,
			&reviewer.LastReviewPageScrapped, &reviewer.LastMoviePageScrapped, &lastScrappedAt,
			&reviewer.CreatedAt, &reviewer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}

		if lastScrappedAt.Valid {
			reviewer.LastScrappedAt = &lastScrappedAt.Time
		}

		reviewers = append(reviewers, &reviewer)
	}

	return reviewers, rows.Err()
}

// ===== PLAYER REVIEWER LINK REPOSITORY =====

// PlayerReviewerLinkRepositoryImpl implémente domain.PlayerReviewerLinkRepository
type PlayerReviewerLinkRepositoryImpl struct {
	db *sql.DB
}

// NewPlayerReviewerLinkRepository crée une nouvelle instance
func NewPlayerReviewerLinkRepository(db *sql.DB) *PlayerReviewerLinkRepositoryImpl {
	return &PlayerReviewerLinkRepositoryImpl{db: db}
}

// Create crée un lien
func (r *PlayerReviewerLinkRepositoryImpl) Create(link *domain.PlayerReviewerLink) error {
	query := `
		INSERT INTO player_reviewer_links (player_id, reviewer_id, is_primary, linked_at)
		VALUES (?, ?, ?, ?)
	`

	now := time.Now()
	link.LinkedAt = now

	_, err := r.db.Exec(query, link.PlayerID, link.ReviewerID, link.IsPrimary, now)
	if err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}

	return nil
}

// Get récupère un lien
func (r *PlayerReviewerLinkRepositoryImpl) Get(playerID, reviewerID string) (*domain.PlayerReviewerLink, error) {
	query := `
		SELECT player_id, reviewer_id, is_primary, linked_at
		FROM player_reviewer_links WHERE player_id = ? AND reviewer_id = ?
	`

	var link domain.PlayerReviewerLink
	err := r.db.QueryRow(query, playerID, reviewerID).Scan(
		&link.PlayerID, &link.ReviewerID, &link.IsPrimary, &link.LinkedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// GetPlayerReviewers récupère tous les reviewers d'un joueur
func (r *PlayerReviewerLinkRepositoryImpl) GetPlayerReviewers(playerID string) ([]*domain.LetterboxdReviewer, error) {
	query := `
		SELECT lr.id, lr.letterboxd_username, lr.total_reviews, lr.total_movies_watched, 
		       lr.last_review_page_scrapped, lr.last_movie_page_scrapped, lr.last_scrapped_at, lr.created_at, lr.updated_at
		FROM letterboxd_reviewers lr
		JOIN player_reviewer_links prl ON lr.id = prl.reviewer_id
		WHERE prl.player_id = ?
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []*domain.LetterboxdReviewer
	for rows.Next() {
		var reviewer domain.LetterboxdReviewer
		var lastScrappedAt sql.NullTime

		err := rows.Scan(
			&reviewer.ID, &reviewer.LetterboxdUsername, &reviewer.TotalReviews, &reviewer.TotalMoviesWatched,
			&reviewer.LastReviewPageScrapped, &reviewer.LastMoviePageScrapped, &lastScrappedAt,
			&reviewer.CreatedAt, &reviewer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}

		if lastScrappedAt.Valid {
			reviewer.LastScrappedAt = &lastScrappedAt.Time
		}

		reviewers = append(reviewers, &reviewer)
	}

	return reviewers, rows.Err()
}

// Delete supprime un lien
func (r *PlayerReviewerLinkRepositoryImpl) Delete(playerID, reviewerID string) error {
	query := `DELETE FROM player_reviewer_links WHERE player_id = ? AND reviewer_id = ?`

	_, err := r.db.Exec(query, playerID, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}
