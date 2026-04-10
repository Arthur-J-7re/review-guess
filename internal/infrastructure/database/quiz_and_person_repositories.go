package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"review-guess/internal/domain"
)

// ===== QUIZ HISTORY REPOSITORY =====

// QuizHistoryRepository implémente domain.QuizHistoryRepository
type QuizHistoryRepositoryImpl struct {
	db *sql.DB
}

// NewQuizHistoryRepository crée une nouvelle instance
func NewQuizHistoryRepository(db *sql.DB) *QuizHistoryRepositoryImpl {
	return &QuizHistoryRepositoryImpl{db: db}
}

// Create crée une nouvelle entrée d'historique
func (r *QuizHistoryRepositoryImpl) Create(answer *domain.QuizAnswer) error {
	// Convert options to JSON
	optionsJSON, err := json.Marshal(answer.Options)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	query := `
		INSERT INTO quiz_history (id, player_id, review_id, correct_movie_id, player_answer_id, is_correct, options, answered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	answer.AnsweredAt = now

	_, err = r.db.Exec(
		query,
		answer.ID, answer.PlayerID, answer.ReviewID, answer.CorrectMovieID,
		answer.PlayerAnswerID, answer.IsCorrect, string(optionsJSON), now,
	)

	if err != nil {
		return fmt.Errorf("failed to create quiz history: %w", err)
	}

	return nil
}

// GetPlayerHistory récupère l'historique total d'un joueur
func (r *QuizHistoryRepositoryImpl) GetPlayerHistory(playerID string) ([]*domain.QuizAnswer, error) {
	query := `
		SELECT id, player_id, review_id, correct_movie_id, player_answer_id, is_correct, options, answered_at
		FROM quiz_history WHERE player_id = ? ORDER BY answered_at DESC
	`

	rows, err := r.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var history []*domain.QuizAnswer
	for rows.Next() {
		var answer domain.QuizAnswer
		var optionsJSON string
		var playerAnswerID sql.NullString

		err := rows.Scan(
			&answer.ID, &answer.PlayerID, &answer.ReviewID, &answer.CorrectMovieID,
			&playerAnswerID, &answer.IsCorrect, &optionsJSON, &answer.AnsweredAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}

		if playerAnswerID.Valid {
			answer.PlayerAnswerID = &playerAnswerID.String
		}

		// Unmarshal options
		err = json.Unmarshal([]byte(optionsJSON), &answer.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal options: %w", err)
		}

		history = append(history, &answer)
	}

	return history, rows.Err()
}

// GetPlayerScores calcule les scores du joueur
func (r *QuizHistoryRepositoryImpl) GetPlayerScores(playerID string) (totalAnswered int, correctAnswers int, err error) {
	query := `SELECT COUNT(*), SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) FROM quiz_history WHERE player_id = ?`

	var correct sql.NullInt64
	err = r.db.QueryRow(query, playerID).Scan(&totalAnswered, &correct)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to get player scores: %w", err)
	}

	if correct.Valid {
		correctAnswers = int(correct.Int64)
	}

	return totalAnswered, correctAnswers, nil
}

// ===== PERSON REPOSITORY =====

// PersonRepository implémente domain.PersonRepository
type PersonRepositoryImpl struct {
	db *sql.DB
}

// NewPersonRepository crée une nouvelle instance
func NewPersonRepository(db *sql.DB) *PersonRepositoryImpl {
	return &PersonRepositoryImpl{db: db}
}

// Get récupère une personne par ID
func (r *PersonRepositoryImpl) Get(id string) (*domain.Person, error) {
	query := `
		SELECT id, tmdb_id, name, profile_picture, created_at FROM people WHERE id = ?
	`

	var person domain.Person
	err := r.db.QueryRow(query, id).Scan(
		&person.ID, &person.TmdbID, &person.Name, &person.ProfilePicture, &person.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get person: %w", err)
	}

	return &person, nil
}

// GetByTmdbID récupère une personne par son ID TMDB
func (r *PersonRepositoryImpl) GetByTmdbID(tmdbID int) (*domain.Person, error) {
	query := `
		SELECT id, tmdb_id, name, profile_picture, created_at FROM people WHERE tmdb_id = ?
	`

	var person domain.Person
	err := r.db.QueryRow(query, tmdbID).Scan(
		&person.ID, &person.TmdbID, &person.Name, &person.ProfilePicture, &person.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get person by TMDB ID: %w", err)
	}

	return &person, nil
}

// Create crée une nouvelle personne
func (r *PersonRepositoryImpl) Create(person *domain.Person) error {
	query := `
		INSERT INTO people (id, tmdb_id, name, profile_picture, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	now := time.Now()
	person.CreatedAt = now

	_, err := r.db.Exec(
		query, person.ID, person.TmdbID, person.Name, person.ProfilePicture, now,
	)

	if err != nil {
		return fmt.Errorf("failed to create person: %w", err)
	}

	return nil
}

// GetMovieCast récupère le casting d'un film
func (r *PersonRepositoryImpl) GetMovieCast(movieID string) ([]*domain.Person, error) {
	query := `
		SELECT p.id, p.tmdb_id, p.name, p.profile_picture, mc.character_name, mc.role_order, p.created_at
		FROM people p
		JOIN movie_cast mc ON p.id = mc.person_id
		WHERE mc.movie_id = ?
		ORDER BY mc.role_order ASC
	`

	rows, err := r.db.Query(query, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cast: %w", err)
	}
	defer rows.Close()

	var cast []*domain.Person
	for rows.Next() {
		var person domain.Person
		err := rows.Scan(
			&person.ID, &person.TmdbID, &person.Name, &person.ProfilePicture,
			&person.CharacterName, &person.RoleOrder, &person.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan person: %w", err)
		}
		cast = append(cast, &person)
	}

	return cast, rows.Err()
}

// GetMovieDirectors récupère les réalisateurs d'un film
func (r *PersonRepositoryImpl) GetMovieDirectors(movieID string) ([]*domain.Person, error) {
	query := `
		SELECT p.id, p.tmdb_id, p.name, p.profile_picture, p.created_at
		FROM people p
		JOIN movie_crew mc ON p.id = mc.person_id
		WHERE mc.movie_id = ? AND mc.job = 'Director'
	`

	rows, err := r.db.Query(query, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to query directors: %w", err)
	}
	defer rows.Close()

	var directors []*domain.Person
	for rows.Next() {
		var person domain.Person
		err := rows.Scan(
			&person.ID, &person.TmdbID, &person.Name, &person.ProfilePicture, &person.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan director: %w", err)
		}
		directors = append(directors, &person)
	}

	return directors, rows.Err()
}
