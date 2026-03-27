package application

import (
	"fmt"
	"os"
	"strings"

	"review-guess/internal/domain"
	"review-guess/internal/infrastructure/scrapper"

	"github.com/charmbracelet/log"
)

// GameService gère la logique du jeu
type GameService struct {
	scrapper domain.ReviewProvider
	state    *domain.GameState
	logger   *log.Logger
}

// NewGameService crée un nouveau service de jeu
func NewGameService(provider domain.ReviewProvider) *GameService {
	return &GameService{
		scrapper: provider,
		logger:   log.New(os.Stderr),
	}
}

// LoadGame charge les données des utilisateurs et prépare le jeu
func (s *GameService) LoadGame(usernames []string, questionCount int) error {
	if len(usernames) == 0 {
		return domain.ErrInvalidUsername
	}

	if questionCount <= 0 {
		return fmt.Errorf("question count must be > 0")
	}

	s.state = &domain.GameState{
		Users:      []*domain.User{},
		AllReviews: []*domain.Review{},
		Questions:  []*domain.Question{},
		Answers:    []*domain.Answer{},
		Score:      0,
		CurrentIdx: 0,
	}

	// Fetch reviews pour chaque utilisateur
	for _, username := range usernames {
		s.logger.Info("Fetching reviews", "user", username)

		reviews, err := s.scrapper.FetchUserReviews(username)
		if err != nil {
			s.logger.Error("Failed to fetch reviews", "user", username, "err", err)
			continue // Continue même si un utilisateur échoue
		}

		// Filtre les bonnes reviews
		reviews = scrapper.FilterQualityReviews(reviews)

		// Ajoute l'utilisateur
		user := &domain.User{
			Username: username,
			Reviews:  reviews,
		}
		s.state.Users = append(s.state.Users, user)
		s.state.AllReviews = append(s.state.AllReviews, reviews...)
	}

	if len(s.state.AllReviews) == 0 {
		return domain.ErrNotEnoughReviews
	}

	if len(s.state.AllReviews) < questionCount {
		s.logger.Warn(
			"Fewer reviews than questions",
			"reviews", len(s.state.AllReviews),
			"questions", questionCount,
		)
		questionCount = len(s.state.AllReviews)
	}

	// Mélange les reviews
	scrapper.ShuffleReviews(s.state.AllReviews)

	// Génère les questions
	for i := 0; i < questionCount; i++ {
		if i >= len(s.state.AllReviews) {
			break
		}

		review := s.state.AllReviews[i]
		question := &domain.Question{
			ReviewIndex: i,
			Review:      review,
			Difficulty:  scrapper.CalculateDifficulty(review),
		}
		s.state.Questions = append(s.state.Questions, question)
	}

	s.logger.Info(
		"Game loaded",
		"users", len(s.state.Users),
		"reviews", len(s.state.AllReviews),
		"questions", len(s.state.Questions),
	)

	return nil
}

// GetCurrentQuestion retourne la question courante (avec review masquée)
func (s *GameService) GetCurrentQuestion() (*domain.Question, error) {
	if s.state == nil {
		return nil, domain.ErrGameNotStarted
	}

	if s.IsGameOver() {
		return nil, domain.ErrGameOver
	}

	return s.state.Questions[s.state.CurrentIdx], nil
}

// SubmitAnswer valide la réponse et attribue les points
func (s *GameService) SubmitAnswer(guessedUser, guessedFilm string) (*domain.Answer, error) {
	if s.state == nil {
		return nil, domain.ErrGameNotStarted
	}

	if s.IsGameOver() {
		return nil, domain.ErrGameOver
	}

	question := s.state.Questions[s.state.CurrentIdx]
	review := question.Review

	// Valide les réponses
	isCorrectUser := normalizeUsername(guessedUser) == review.Author
	isCorrectFilm := normalizeFilmSlug(guessedFilm) == review.Film.Slug

	// Calcule les points
	points := 0
	if isCorrectUser && isCorrectFilm {
		points = 100
	} else if isCorrectUser {
		points = 50
	} else if isCorrectFilm {
		points = 50
	}

	answer := &domain.Answer{
		QuestionIdx:   s.state.CurrentIdx,
		GuessedUser:   guessedUser,
		GuessedFilm:   guessedFilm,
		IsCorrectUser: isCorrectUser,
		IsCorrectFilm: isCorrectFilm,
		Points:        points,
	}

	s.state.Answers = append(s.state.Answers, answer)
	s.state.Score += points
	s.state.CurrentIdx++

	return answer, nil
}

// GetScore retourne le score courant
func (s *GameService) GetScore() int {
	if s.state == nil {
		return 0
	}
	return s.state.Score
}

// IsGameOver vérifie si le jeu est terminé
func (s *GameService) IsGameOver() bool {
	if s.state == nil {
		return true
	}
	return s.state.CurrentIdx >= len(s.state.Questions)
}

// GetResults retourne les résultats finaux
func (s *GameService) GetResults() *domain.GameResults {
	if s.state == nil {
		return nil
	}

	totalPoints := len(s.state.Questions) * 100
	percentage := float32(0)
	if totalPoints > 0 {
		percentage = (float32(s.state.Score) / float32(totalPoints)) * 100
	}

	return &domain.GameResults{
		TotalScore:  s.state.Score,
		TotalPoints: totalPoints,
		Percentage:  percentage,
		Answers:     s.state.Answers,
	}
}

// normalizeUsername normalise un username pour comparaison
func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// normalizeFilmSlug normalise un film slug pour comparaison
func normalizeFilmSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
