package application

import (
	"testing"

	"review-guess/internal/domain"
)

// MockReviewProvider implémente domain.ReviewProvider pour les tests
type MockReviewProvider struct {
	reviews map[string][]*domain.Review
}

func NewMockReviewProvider() *MockReviewProvider {
	return &MockReviewProvider{
		reviews: make(map[string][]*domain.Review),
	}
}

func (m *MockReviewProvider) FetchUserReviews(username string) ([]*domain.Review, error) {
	if reviews, ok := m.reviews[username]; ok {
		return reviews, nil
	}
	return []*domain.Review{}, nil
}

func (m *MockReviewProvider) AddReview(username string, review *domain.Review) {
	m.reviews[username] = append(m.reviews[username], review)
}

// Tests
func TestGameServiceLoadGame(t *testing.T) {
	provider := NewMockReviewProvider()

	// Ajoute des reviews de test
	film1 := &domain.Film{Slug: "godfather-1972", Title: "The Godfather"}
	film2 := &domain.Film{Slug: "pulp-fiction-1994", Title: "Pulp Fiction"}

	provider.AddReview("alice", &domain.Review{
		ID:      "alice-godfather-1972",
		Author:  "alice",
		Film:    film1,
		Content: "This is a really great movie with excellent cinematography",
		Rating:  5,
		Liked:   true,
	})

	provider.AddReview("bob", &domain.Review{
		ID:      "bob-pulp-fiction-1994",
		Author:  "bob",
		Film:    film2,
		Content: "Amazing dialogue and direction",
		Rating:  5,
		Liked:   true,
	})

	service := NewGameService(provider)
	err := service.LoadGame([]string{"alice", "bob"}, 2)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(service.state.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(service.state.Users))
	}

	if len(service.state.Questions) != 2 {
		t.Errorf("expected 2 questions, got %d", len(service.state.Questions))
	}
}

func TestGameServiceSubmitAnswer(t *testing.T) {
	provider := NewMockReviewProvider()

	film := &domain.Film{Slug: "godfather-1972", Title: "The Godfather"}
	provider.AddReview("alice", &domain.Review{
		ID:      "alice-godfather-1972",
		Author:  "alice",
		Film:    film,
		Content: "This is a really great movie",
		Rating:  5,
		Liked:   true,
	})

	service := NewGameService(provider)
	service.LoadGame([]string{"alice"}, 1)

	// Réponse correcte
	answer, err := service.SubmitAnswer("alice", "godfather-1972")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !answer.IsCorrectUser || !answer.IsCorrectFilm {
		t.Errorf("expected both answers to be correct")
	}

	if answer.Points != 100 {
		t.Errorf("expected 100 points for correct answer, got %d", answer.Points)
	}

	if service.GetScore() != 100 {
		t.Errorf("expected score 100, got %d", service.GetScore())
	}
}

func TestGameServicePartialAnswer(t *testing.T) {
	provider := NewMockReviewProvider()

	film := &domain.Film{Slug: "godfather-1972", Title: "The Godfather"}
	provider.AddReview("alice", &domain.Review{
		ID:      "alice-godfather-1972",
		Author:  "alice",
		Film:    film,
		Content: "This is a really great movie",
		Rating:  5,
	})

	service := NewGameService(provider)
	service.LoadGame([]string{"alice"}, 1)

	// Réponse partiellement correcte (film correct, user mal)
	answer, err := service.SubmitAnswer("bob", "godfather-1972")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if answer.IsCorrectUser {
		t.Errorf("expected user to be wrong")
	}

	if !answer.IsCorrectFilm {
		t.Errorf("expected film to be correct")
	}

	if answer.Points != 50 {
		t.Errorf("expected 50 points for partial answer, got %d", answer.Points)
	}
}

func TestGameServiceGameOver(t *testing.T) {
	provider := NewMockReviewProvider()

	film := &domain.Film{Slug: "godfather-1972", Title: "The Godfather"}
	provider.AddReview("alice", &domain.Review{
		ID:      "alice-godfather-1972",
		Author:  "alice",
		Film:    film,
		Content: "This is a really great movie",
		Rating:  5,
	})

	service := NewGameService(provider)
	service.LoadGame([]string{"alice"}, 1)

	if service.IsGameOver() {
		t.Errorf("game should not be over before answering")
	}

	service.SubmitAnswer("alice", "godfather-1972")

	if !service.IsGameOver() {
		t.Errorf("game should be over after answering all questions")
	}
}

func TestGameServiceResults(t *testing.T) {
	provider := NewMockReviewProvider()

	film := &domain.Film{Slug: "godfather-1972", Title: "The Godfather"}
	provider.AddReview("alice", &domain.Review{
		ID:      "alice-godfather-1972",
		Author:  "alice",
		Film:    film,
		Content: "This is a really great movie",
		Rating:  5,
	})

	service := NewGameService(provider)
	service.LoadGame([]string{"alice"}, 1)
	service.SubmitAnswer("alice", "godfather-1972")

	results := service.GetResults()
	if results == nil {
		t.Errorf("expected non-nil results")
	}

	if results.TotalScore != 100 {
		t.Errorf("expected total score 100, got %d", results.TotalScore)
	}

	if results.TotalPoints != 100 {
		t.Errorf("expected total points 100, got %d", results.TotalPoints)
	}

	if results.Percentage != 100 {
		t.Errorf("expected 100%%, got %.1f%%", results.Percentage)
	}
}
