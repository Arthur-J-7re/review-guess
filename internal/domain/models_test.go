package domain

import (
	"testing"
)

func TestUserCreation(t *testing.T) {
	user := &User{
		Username: "john_doe",
		Name:     "John Doe",
		Reviews:  []*Review{},
	}

	if user.Username != "john_doe" {
		t.Errorf("expected username john_doe, got %s", user.Username)
	}
	if len(user.Reviews) != 0 {
		t.Errorf("expected 0 reviews, got %d", len(user.Reviews))
	}
}

func TestFilmCreation(t *testing.T) {
	film := &Film{
		Slug:      "the-godfather-1972",
		Title:     "The Godfather",
		Year:      1972,
		Directors: []string{"Francis Ford Coppola"},
		Poster:    "https://example.com/poster.jpg",
	}

	if film.Title != "The Godfather" {
		t.Errorf("expected title The Godfather, got %s", film.Title)
	}
	if film.Year != 1972 {
		t.Errorf("expected year 1972, got %d", film.Year)
	}
}

func TestReviewCreation(t *testing.T) {
	film := &Film{
		Slug:  "pulp-fiction-1994",
		Title: "Pulp Fiction",
		Year:  1994,
	}

	review := &Review{
		ID:       "alice-pulp-fiction-1994",
		Author:   "alice",
		Film:     film,
		Content:  "Amazing movie!",
		Rating:   5,
		Liked:    true,
		Spoilers: false,
	}

	if review.Author != "alice" {
		t.Errorf("expected author alice, got %s", review.Author)
	}
	if review.Rating != 5 {
		t.Errorf("expected rating 5, got %d", review.Rating)
	}
	if !review.Liked {
		t.Errorf("expected review to be liked")
	}
}

func TestGameStateInitialization(t *testing.T) {
	state := &GameState{
		Users:      []*User{},
		AllReviews: []*Review{},
		Questions:  []*Question{},
		Answers:    []*Answer{},
		Score:      0,
		CurrentIdx: 0,
	}

	if state.Score != 0 {
		t.Errorf("expected initial score 0, got %d", state.Score)
	}
	if state.CurrentIdx != 0 {
		t.Errorf("expected initial index 0, got %d", state.CurrentIdx)
	}
}

func TestAnswerCreation(t *testing.T) {
	answer := &Answer{
		QuestionIdx:   0,
		GuessedUser:   "bob",
		GuessedFilm:   "the-godfather-1972",
		IsCorrectUser: true,
		IsCorrectFilm: true,
		Points:        100,
	}

	if answer.Points != 100 {
		t.Errorf("expected 100 points, got %d", answer.Points)
	}
	if !answer.IsCorrectUser || !answer.IsCorrectFilm {
		t.Errorf("expected both answers to be correct")
	}
}

func TestGameResults(t *testing.T) {
	results := &GameResults{
		TotalScore:  750,
		TotalPoints: 1000,
		Percentage:  75.0,
		Answers:     []*Answer{},
	}

	if results.Percentage != 75.0 {
		t.Errorf("expected 75%%, got %.1f%%", results.Percentage)
	}
}
