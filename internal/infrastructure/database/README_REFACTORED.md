# Database Infrastructure Setup (REFACTORED)

This directory contains the database layer for the quiz system with **Players vs Reviewers** architecture.

## Structure

```
internal/infrastructure/database/
├── client.go                          # SQLite client with auto-migrations
├── movie_repository.go                # Movie CRUD + GetMany
├── player_and_reviewer_repositories.go  # NEW: Player, Reviewer, Links CRUD
├── review_repository.go               # Review CRUD (uses reviewer_id)
├── other_repositories.go              # Similarity + ReviewerMovie repos
├── quiz_and_person_repositories.go    # QuizHistory + People repos
```

## Initialization

### 1. Add dependency to go.mod

```bash
go get github.com/mattn/go-sqlite3
```

### 2. Initialize database in main.go

```go
package main

import (
	"log"
	"net/http"

	"review-guess/internal/adapters/httpapi"
	"review-guess/internal/infrastructure/database"
	"review-guess/internal/infrastructure/scrapper"
)

func main() {
	// === DATABASE ===
	db, err := database.NewSQLiteClient("./review-guess.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// === REPOSITORIES ===
	sqlDB := db.GetDB()
	
	// Player & Reviewer management
	playerRepo := database.NewPlayerRepository(sqlDB)
	reviewerRepo := database.NewLetterboxdReviewerRepository(sqlDB)
	linkRepo := database.NewPlayerReviewerLinkRepository(sqlDB)
	
	// Movie & Review data
	movieRepo := database.NewMovieRepository(sqlDB)
	reviewRepo := database.NewReviewRepository(sqlDB)
	reviewerMovieRepo := database.NewReviewerMovieRepository(sqlDB)
	
	// Similarity & Quiz tracking
	similarityRepo := database.NewMovieSimilarityRepository(sqlDB)
	quizRepo := database.NewQuizHistoryRepository(sqlDB)
	personRepo := database.NewPersonRepository(sqlDB)
	
	// === ROUTER ===
	router := httpapi.NewRouter(
		playerRepo, reviewerRepo, linkRepo,
		movieRepo, reviewRepo, reviewerMovieRepo,
		similarityRepo, quizRepo, personRepo,
	)

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

## Repository Usage Examples

### Player Management

```go
playerRepo := database.NewPlayerRepository(sqlDB)

// Anonymous player
player := &domain.Player{
	ID:         "player-anon-123",
	IsLoggedIn: false,
	Nickname:   nil,  // Anonymous
}
err := playerRepo.Create(player)

// Logged-in player
player := &domain.Player{
	ID:         "player-alice-123",
	IsLoggedIn: true,
	Nickname:   "alice",
}
err := playerRepo.Create(player)

// Retrieve
player, err := playerRepo.Get("player-alice-123")

// Update LastPlayedAt
err := playerRepo.Update(player)

// List all players
players, err := playerRepo.List() // Orders by created_at
```

### Letterboxd Reviewer Management

```go
reviewerRepo := database.NewLetterboxdReviewerRepository(sqlDB)

// Create (when scraping discovers new account)
reviewer := &domain.LetterboxdReviewer{
	ID:                   "reviewer-alice",
	LetterboxdUsername:   "alice",
	TotalReviews:         150,
	TotalMoviesWatched:   250,
}
err := reviewerRepo.Create(reviewer)

// Get by username (lookup when scraping)
reviewer, err := reviewerRepo.GetByUsername("alice")

// Get by ID
reviewer, err := reviewerRepo.Get("reviewer-alice")

// Update stats (periodically after scraping)
reviewer.TotalReviews = 151
reviewer.LastScrappedAt = time.Now()
err := reviewerRepo.Update(reviewer)

// List all reviewers
reviewers, err := reviewerRepo.List()
```

### Player-Reviewer Links

```go
linkRepo := database.NewPlayerReviewerLinkRepository(sqlDB)

// Link logged player to their Letterboxd account
link := &domain.PlayerReviewerLink{
	PlayerID:   "player-alice",
	ReviewerID: "reviewer-alice",
	IsPrimary:  true,
}
err := linkRepo.Create(link)

// Get player's reviewers
reviewers, err := linkRepo.GetPlayerReviewers("player-alice")

// Delete link (player unlinks account)
err := linkRepo.Delete("player-alice", "reviewer-alice")

// Get link to check if exists
link, err := linkRepo.Get("player-alice", "reviewer-alice")
```

### Review Management (Tied to Reviewers)

```go
reviewRepo := database.NewReviewRepository(sqlDB)

// Create review (from scraping)
review := &domain.Review{
	ID:         "review-123",
	ReviewerID: "reviewer-alice",  // NOT player
	MovieID:    "tt0111161",
	Title:      "Amazing!",
	Content:    "Best movie ever...",
	Rating:     5.0,
	Usable:     true,  // Can be quiz question
}
err := reviewRepo.Create(review)

// Get by reviewer and movie (check if exists)
review, err := reviewRepo.GetByReviewerAndMovie("reviewer-alice", "tt0111161")

// Get usable reviews from a reviewer (for quiz questions)
reviews, err := reviewRepo.GetUsableReviewsForReviewer("reviewer-alice")

// Get random usable review (anonymous player + random reviewer)
randomReview, err := reviewRepo.GetRandomUsableReview()

// Batch create (during scraping)
var reviews []*domain.Review
for _, scrapedReview := range scrapedData {
	reviews = append(reviews, &domain.Review{
		ID:         generateID(),
		ReviewerID: "reviewer-alice",
		MovieID:    scrapedReview.MovieID,
		Content:    scrapedReview.Content,
		Rating:     scrapedReview.Rating,
		Usable:     true,
	})
}
err := reviewRepo.CreateBatch(reviews)  // Single transaction
```

### Reviewer Movie Tracking

```go
reviewerMovieRepo := database.NewReviewerMovieRepository(sqlDB)

// Track film watched by reviewer
reviewerMovie := &domain.ReviewerMovie{
	ReviewerID: "reviewer-alice",
	MovieID:    "tt0111161",
	WatchedAt:  time.Now(),
}
err := reviewerMovieRepo.Create(reviewerMovie)

// Get all films watched by reviewer
watched, err := reviewerMovieRepo.GetMoviesWatchedByReviewer("reviewer-alice")

// Get films NOT watched by reviewer (for lures)
unwatched, err := reviewerMovieRepo.GetMoviesNotWatchedByReviewer("reviewer-alice", 100)

// Batch create (from scraping profile)
var movies []*domain.ReviewerMovie
for _, movieID := range scrapedWatchlist {
	movies = append(movies, &domain.ReviewerMovie{
		ReviewerID: "reviewer-alice",
		MovieID:    movieID,
	})
}
err := reviewerMovieRepo.CreateBatch(movies)
```

### Movie Similarity

```go
similarityRepo := database.NewMovieSimilarityRepository(sqlDB)

// Create similarity
similarity := &domain.MovieSimilarity{
	MovieAID:        "tt0111161",
	MovieBID:        "tt0068646",
	SimilarityScore: 0.85,
	SharedDirectors: 1,
	SharedActors:    2,
	SharedGenres:    2,
	YearProximity:   1,
}
err := similarityRepo.Create(similarity)

// Find most similar films (for lures)
similarities, err := similarityRepo.GetTopSimilarMovies("tt0111161", 4)
var lures []string
for _, sim := range similarities {
	lures = append(lures, sim.MovieBID)
}
```

### Quiz History (Tied to Players)

```go
quizRepo := database.NewQuizHistoryRepository(sqlDB)

// Record quiz answer
answer := &domain.QuizAnswer{
	ID:               "answer-123",
	PlayerID:         "player-alice",  // Track by player
	ReviewID:         "review-456",
	CorrectMovieID:   "tt0111161",
	PlayerAnswerID:   &selectedMovieID,  // NULL if skipped
	IsCorrect:        selectedMovieID == "tt0111161",
	Options:          []string{"tt0111161", "tt0068646", ...},
}
err := quizRepo.Create(answer)

// Get player's quiz history
history, err := quizRepo.GetPlayerHistory("player-alice")

// Get player's scores
total, correct, err := quizRepo.GetPlayerScores("player-alice")
accuracy := float64(correct) / float64(total) * 100
fmt.Printf("Score: %d/%d (%.1f%%)\n", correct, total, accuracy)
```

## Key Architectural Points

### Anonymous Player Flow
```go
// 1. Create anonymous player
player := &domain.Player{
	ID:         generateID(),
	IsLoggedIn: false,
}
playerRepo.Create(player)

// 2. Get random review from random reviewer
review, err := reviewRepo.GetRandomUsableReview()

// 3. Record answer (player doesn't link to reviewer)
answer := &domain.QuizAnswer{
	PlayerID: player.ID,
	ReviewID: review.ID,
	...
}
quizRepo.Create(answer)
```

### Logged Player Flow
```go
// 1. Create logged player
player := &domain.Player{
	ID:         generateID(),
	IsLoggedIn: true,
	Nickname:   "alice",
}
playerRepo.Create(player)

// 2. Link to their Letterboxd account
reviewer, _ := reviewerRepo.GetByUsername("alice")
link := &domain.PlayerReviewerLink{
	PlayerID:   player.ID,
	ReviewerID: reviewer.ID,
	IsPrimary:  true,
}
linkRepo.Create(link)

// 3. Get their reviews only
reviews, err := reviewRepo.GetUsableReviewsForReviewer(reviewer.ID)

// 4. Play quiz with their own reviews
```

### Scraper Flow
```go
// 1. Discover new Letterboxd account
reviewer := &domain.LetterboxdReviewer{
	ID:                 "reviewer-bob",
	LetterboxdUsername: "bob",
}
reviewerRepo.Create(reviewer)

// 2. Scrape and store reviews
reviews := []*domain.Review{...}
reviewRepo.CreateBatch(reviews)

// 3. Track watched movies
movies := []*domain.ReviewerMovie{...}
reviewerMovieRepo.CreateBatch(movies)

// Note: No players involved - just data collection
```

## Transactions

For multi-table operations:

```go
tx, err := db.BeginTx()
if err != nil {
	log.Fatal(err)
}
defer tx.Rollback()

// Multiple operations...
playerRepo.Create(player)   // Uses tx
reviewerRepo.Create(reviewer)
linkRepo.Create(link)

if err := tx.Commit(); err != nil {
	log.Fatal(err)
}
```

## Performance Indexes

Automatically created during initialization:

- `idx_players_created_at` - Player signup queries
- `idx_players_last_played_at` - Recent activity
- `idx_letterboxd_reviewers_updated_at` - Scraping order
- `idx_reviews_reviewer_id` - Reviews by source
- `idx_reviews_usable` - Quiz questions
- `idx_reviewer_movies_reviewer_id` - Films watched
- `idx_movie_similarities_movie_a` - Similar films
- `idx_quiz_history_player_id` - Player scores

## Breaking Changes from Old Architecture

| Old | New | Reason |
|-----|-----|--------|
| `User` model | `Player` + `LetterboxdReviewer` | Separate concerns |
| `UserRepository` | `PlayerRepository` | Per-entity repository |
| `reviews.user_id` | `reviews.reviewer_id` | Reviews come from Letterboxd |
| `user_movies` table | `reviewer_movies` | Track reviewer's watchlist |
| `quiz_history.user_id` | `quiz_history.player_id` | Quiz is per-player |
| `UserMovieRepository` | `ReviewerMovieRepository` | Clear data ownership |
| `GetUsableReviewsForUser()` | `GetUsableReviewsForReviewer()` | Method clarity |
| `GetUserHistory()` | `GetPlayerHistory()` | Method clarity |

## Documentation

- **DATABASE_REFACTORED.md** - Complete schema reference
- **ARCHITECTURE_REFACTORED.md** - Design patterns and use cases
- **REFACTORING_SUMMARY.md** - Migration guide from old architecture
