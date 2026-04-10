# Implementation Checklist - Phase 4: Integration

After refactoring to Player/Reviewer architecture, integrate the database into your running system.

## ✅ Step 1: Add go-sqlite3 Dependency

**Required before anything works**

```bash
go get github.com/mattn/go-sqlite3
go mod tidy
```

**Verify:**
```bash
grep mattn go.mod
# Should show: github.com/mattn/go-sqlite3 v1.14.x (or similar)
```

---

## ✅ Step 2: Update cmd/review-guess/main.go

Add database initialization and repository injection.

**File:** `cmd/review-guess/main.go`

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
	// === Step 1: Initialize Database ===
	db, err := database.NewSQLiteClient("./review-guess.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// === Step 2: Create All Repositories ===
	sqlDB := db.GetDB()
	
	playerRepo := database.NewPlayerRepository(sqlDB)
	reviewerRepo := database.NewLetterboxdReviewerRepository(sqlDB)
	linkRepo := database.NewPlayerReviewerLinkRepository(sqlDB)
	
	movieRepo := database.NewMovieRepository(sqlDB)
	reviewRepo := database.NewReviewRepository(sqlDB)
	reviewerMovieRepo := database.NewReviewerMovieRepository(sqlDB)
	
	similarityRepo := database.NewMovieSimilarityRepository(sqlDB)
	quizRepo := database.NewQuizHistoryRepository(sqlDB)
	personRepo := database.NewPersonRepository(sqlDB)

	// === Step 3: Initialize Scraper & Services ===
	scraper := scrapper.NewLetterboxdScraper()
	
	// === Step 4: Create Router with Repos ===
	router := httpapi.NewRouter(
		playerRepo, reviewerRepo, linkRepo,
		movieRepo, reviewRepo, reviewerMovieRepo,
		similarityRepo, quizRepo, personRepo,
		scraper,
	)

	// === Step 5: Start Server ===
	log.Println("🎬 Review-Guess server starting on :8080")
	log.Printf("📁 Database: ./review-guess.db\n")
	log.Printf("🔗 API: http://localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

**Verify:**
```bash
cd cmd/review-guess
go run main.go
# Should see: "Review-Guess server starting on :8080"
# Should create ./review-guess.db automatically
```

---

## ✅ Step 3: Update HTTP Router

**File:** `internal/adapters/httpapi/router.go`

Update constructor to accept all repositories:

```go
package httpapi

import (
	"net/http"
	"review-guess/internal/domain"
	"github.com/gorilla/mux"
)

type Router struct {
	mux *mux.Router
	
	// Player & Reviewer
	playerRepo    domain.PlayerRepository
	reviewerRepo  domain.LetterboxdReviewerRepository
	linkRepo      domain.PlayerReviewerLinkRepository
	
	// Data
	movieRepo     domain.MovieRepository
	reviewRepo    domain.ReviewRepository
	reviewerMovieRepo domain.ReviewerMovieRepository
	
	// Similarity & Quiz
	similarityRepo domain.MovieSimilarityRepository
	quizRepo      domain.QuizHistoryRepository
	personRepo    domain.PersonRepository
}

func NewRouter(
	playerRepo domain.PlayerRepository,
	reviewerRepo domain.LetterboxdReviewerRepository,
	linkRepo domain.PlayerReviewerLinkRepository,
	movieRepo domain.MovieRepository,
	reviewRepo domain.ReviewRepository,
	reviewerMovieRepo domain.ReviewerMovieRepository,
	similarityRepo domain.MovieSimilarityRepository,
	quizRepo domain.QuizHistoryRepository,
	personRepo domain.PersonRepository,
	scraper *scrapper.LetterboxdScraper,
) *Router {
	r := mux.NewRouter()
	
	router := &Router{
		mux:               r,
		playerRepo:        playerRepo,
		reviewerRepo:      reviewerRepo,
		linkRepo:          linkRepo,
		movieRepo:         movieRepo,
		reviewRepo:        reviewRepo,
		reviewerMovieRepo: reviewerMovieRepo,
		similarityRepo:    similarityRepo,
		quizRepo:          quizRepo,
		personRepo:        personRepo,
	}
	
	// Register endpoints
	router.setupRoutes()
	
	return router
}

func (ro *Router) setupRoutes() {
	// Health check (existing)
	ro.mux.HandleFunc("/health", ro.healthHandler).Methods("GET")
	
	// Quiz endpoints (NEW - from Phase 4)
	ro.mux.HandleFunc("/api/quiz/next", ro.getQuizQuestion).Methods("GET")
	ro.mux.HandleFunc("/api/quiz/answer", ro.recordQuizAnswer).Methods("POST")
	ro.mux.HandleFunc("/api/quiz/stats", ro.getPlayerStats).Methods("GET")
	
	// Reviewer endpoints (NEW)
	ro.mux.HandleFunc("/api/reviewers", ro.listReviewers).Methods("GET")
	ro.mux.HandleFunc("/api/reviewers/{username}", ro.getReviewerByUsername).Methods("GET")
}

func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ro.mux.ServeHTTP(w, r)
}
```

---

## ✅ Step 4: Create Quiz Endpoints

**File:** `internal/adapters/httpapi/handlers.go`

```go
package httpapi

import (
	"encoding/json"
	"net/http"
	"review-guess/internal/domain"
)

// GET /api/quiz/next?player_id=xxx&reviewer_id=yyy
func (ro *Router) getQuizQuestion(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	reviewerID := r.URL.Query().Get("reviewer_id")
	
	if playerID == "" {
		http.Error(w, "player_id required", http.StatusBadRequest)
		return
	}
	
	var review *domain.Review
	var err error
	
	if reviewerID != "" {
		// Specific reviewer
		review, err = ro.reviewRepo.GetUsableReviewsForReviewer(reviewerID)
	} else {
		// Random reviewer
		review, err = ro.reviewRepo.GetRandomUsableReview()
	}
	
	if err != nil || review == nil {
		http.Error(w, "No quiz questions available", http.StatusNotFound)
		return
	}
	
	// Find similar films for lures
	similarities, _ := ro.similarityRepo.GetTopSimilarMovies(review.MovieID, 4)
	options := make([]string, 0)
	
	// Add correct answer
	options = append(options, review.MovieID)
	
	// Add lures (similar films)
	for _, sim := range similarities {
		options = append(options, sim.MovieBID)
	}
	
	// Shuffle options
	// TODO: Implement shuffle function
	
	response := map[string]interface{}{
		"review_id": review.ID,
		"title":     review.Title,
		"content":   review.Content,
		"options":   options,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// POST /api/quiz/answer
func (ro *Router) recordQuizAnswer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PlayerID   string `json:"player_id"`
		ReviewID   string `json:"review_id"`
		AnswerID   string `json:"answer_id"` // Movie selected by player
	}
	
	json.NewDecoder(r.Body).Decode(&req)
	
	if req.PlayerID == "" || req.ReviewID == "" {
		http.Error(w, "player_id and review_id required", http.StatusBadRequest)
		return
	}
	
	// Get review to know correct answer
	review, err := ro.reviewRepo.Get(req.ReviewID)
	if err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}
	
	// Get all options that were shown
	// TODO: Store options in quiz_history.options or pass from client
	
	isCorrect := req.AnswerID == review.MovieID
	
	answer := &domain.QuizAnswer{
		ID:               generateID(),
		PlayerID:         req.PlayerID,
		ReviewID:         req.ReviewID,
		CorrectMovieID:   review.MovieID,
		PlayerAnswerID:   &req.AnswerID,
		IsCorrect:        isCorrect,
		// Options:          options,  // TODO
	}
	
	err = ro.quizRepo.Create(answer)
	if err != nil {
		http.Error(w, "Failed to record answer", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"correct":        isCorrect,
		"correct_movie":  review.MovieID,
	})
}

// GET /api/quiz/stats?player_id=xxx
func (ro *Router) getPlayerStats(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	
	if playerID == "" {
		http.Error(w, "player_id required", http.StatusBadRequest)
		return
	}
	
	total, correct, err := ro.quizRepo.GetPlayerScores(playerID)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}
	
	accuracy := 0.0
	if total > 0 {
		accuracy = float64(correct) / float64(total) * 100
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"player_id": playerID,
		"total":     total,
		"correct":   correct,
		"accuracy":  accuracy,
	})
}

// Helper endpoints
func (ro *Router) listReviewers(w http.ResponseWriter, r *http.Request) {
	reviewers, _ := ro.reviewerRepo.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewers)
}

func (ro *Router) getReviewerByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username") // gorilla/mux: use mux.Vars(r)
	reviewer, err := ro.reviewerRepo.GetByUsername(username)
	if err != nil {
		http.Error(w, "Reviewer not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewer)
}

func (ro *Router) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "review-guess",
	})
}
```

---

## ✅ Step 5: Build and Test

```bash
# Build
cd cmd/review-guess
go build -o review-guess .

# Run
./review-guess

# In another terminal:

# Test health
curl http://localhost:8080/health

# Test reviewers list
curl http://localhost:8080/api/reviewers

# Test quiz (with sample data)
curl "http://localhost:8080/api/quiz/next?player_id=test-player-1"

# Test stats
curl "http://localhost:8080/api/quiz/stats?player_id=test-player-1"
```

---

## 📋 Before You Go To Production

- [ ] Run migrations manually to verify: `sqlite3 review-guess.db < migrations/001_init_schema.sql`
- [ ] Populate test data (players, reviewers, movies, reviews) via API or scripts
- [ ] Implement MovieMatcher similarity calculator (`application/movie_matcher.go`)
- [ ] Implement pre-calculation batch job for similarities
- [ ] Add TMDB enrichment for better movie data
- [ ] Add scraper integration to populate reviewers and reviews
- [ ] Implement error handling and validation in handlers
- [ ] Add logging and monitoring
- [ ] Add CORS headers if needed for frontend
- [ ] Set up database backups

---

## 🔗 Related Documentation

- **DATABASE_REFACTORED.md** - Complete schema reference
- **ARCHITECTURE_REFACTORED.md** - Design patterns
- **REFACTORING_SUMMARY.md** - Migration guide
- **internal/infrastructure/database/README_REFACTORED.md** - Repository usage

---

## 🆘 Troubleshooting

**Error: "cannot find package github.com/mattn/go-sqlite3"**
```bash
go get github.com/mattn/go-sqlite3
go mod tidy
```

**Error: "table players already exists"**
- Database file already exists with old schema
- Delete `review-guess.db` and restart (auto-migrates)
- Or update migrations manually for schema changes

**Error: "UserRepository not found"**
- Old code still references removed `UserRepository`
- Update to `PlayerRepository` or `LetterboxdReviewerRepository`
- See REFACTORING_SUMMARY.md for all changes

**No quiz questions returned**
- Check if reviews marked as `usable=true`
- Check if reviewer has any reviews with associated movies
- Verify movies in database: `SELECT * FROM movies LIMIT 5`
