package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	"review-guess/internal/application"
	"review-guess/internal/domain"
	"review-guess/internal/infrastructure/scrapper"
)

// Router gère les routes HTTP
type Router struct {
	// Repositories
	playerRepo        domain.PlayerRepository
	reviewerRepo      domain.LetterboxdReviewerRepository
	linkRepo          domain.PlayerReviewerLinkRepository
	movieRepo         domain.MovieRepository
	reviewRepo        domain.ReviewRepository
	reviewerMovieRepo domain.ReviewerMovieRepository
	similarityRepo    domain.MovieSimilarityRepository
	quizRepo          domain.QuizHistoryRepository
	personRepo        domain.PersonRepository
	scraper           *scrapper.Scrapper

	// Services
	reviewService         *application.ReviewService
	reviewerUpdateService *application.ReviewerUpdateService
}

// NewRouter crée un nouveau router avec les repositories
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
	scraper *scrapper.Scrapper,
) *Router {
	reviewService := application.NewReviewService(scraper)
	reviewerUpdateService := application.NewReviewerUpdateService(
		reviewerRepo,
		reviewRepo,
		movieRepo,
		scraper,
	)

	return &Router{
		playerRepo:            playerRepo,
		reviewerRepo:          reviewerRepo,
		linkRepo:              linkRepo,
		movieRepo:             movieRepo,
		reviewRepo:            reviewRepo,
		reviewerMovieRepo:     reviewerMovieRepo,
		similarityRepo:        similarityRepo,
		quizRepo:              quizRepo,
		personRepo:            personRepo,
		scraper:               scraper,
		reviewService:         reviewService,
		reviewerUpdateService: reviewerUpdateService,
	}
}

// Register enregistre toutes les routes
func (r *Router) Register() http.Handler {
	mux := http.NewServeMux()

	// Routes API - Scraper
	mux.HandleFunc("GET /api/reviews", r.handleGetReviews)

	// Routes API - Reviewer management
	mux.HandleFunc("GET /api/reviewers", r.handleListReviewers)
	mux.HandleFunc("GET /api/reviewers/{username}", r.handleGetReviewerByUsername)

	// Routes API - Quiz
	mux.HandleFunc("GET /api/quiz/next", r.handleGetQuizQuestion)
	mux.HandleFunc("POST /api/quiz/answer", r.handleRecordQuizAnswer)
	mux.HandleFunc("GET /api/quiz/stats", r.handleGetPlayerStats)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		jsonResponse(w, http.StatusOK, Response{
			Success: true,
			Data:    "Review Guess API v2.0 - Player/Reviewer Architecture",
		})
	})

	return mux
}

// handleGetReviews traite la requête GET /api/reviews
// Accepte les usernames séparés par des virgules
// Met automatiquement à jour les données du reviewer avant de retourner les reviews
func (r *Router) handleGetReviews(w http.ResponseWriter, req *http.Request) {
	usernamesParam := req.URL.Query().Get("username")
	if usernamesParam == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "username parameter is required (comma-separated)",
		})
		return
	}

	// Split par virgule et trim les espaces
	usernames := []string{}
	for _, username := range strings.Split(usernamesParam, ",") {
		if trimmed := strings.TrimSpace(username); trimmed != "" {
			usernames = append(usernames, trimmed)
		}
	}

	if len(usernames) == 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "at least one valid username is required",
		})
		return
	}

	// 1. Pour chaque username, mettre à jour les données (avec cache token 1h)
	for _, username := range usernames {
		_, err := r.reviewerUpdateService.UpdateReviewerData(username)
		if err != nil {
			// Log l'erreur mais continue avec les autres usernames
			fmt.Printf("Warning: failed to update reviewer %s: %v\n", username, err)
		}
	}

	// 2. Récupérer toutes les reviews des reviewers
	reviews, err := r.reviewService.GetReviews(usernames...)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    reviews,
	})
}

// ===== QUIZ ENDPOINTS =====

// handleGetQuizQuestion returns the next quiz question
// GET /api/quiz/next?player_id=xxx
func (r *Router) handleGetQuizQuestion(w http.ResponseWriter, req *http.Request) {
	playerID := req.URL.Query().Get("player_id")
	if playerID == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "player_id is required",
		})
		return
	}

	// Get a usable review
	review, err := r.reviewRepo.GetRandomUsableReview()
	if err != nil || review == nil {
		jsonResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "No quiz questions available",
		})
		return
	}

	// Get similar movies for lures
	similarities, err := r.similarityRepo.GetTopSimilarMovies(review.MovieID, 4)
	if err != nil {
		options := []string{review.MovieID}
		jsonResponse(w, http.StatusOK, Response{
			Success: true,
			Data: map[string]interface{}{
				"review_id": review.ID,
				"title":     review.Title,
				"content":   review.Content,
				"options":   options,
			},
		})
		return
	}

	// Build options: correct answer + lures
	options := []string{review.MovieID}
	for _, sim := range similarities {
		options = append(options, sim.MovieBID)
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"review_id":        review.ID,
			"correct_movie_id": review.MovieID,
			"title":            review.Title,
			"content":          review.Content,
			"options":          options,
		},
	})
}

// handleRecordQuizAnswer records a player's answer
// POST /api/quiz/answer
func (r *Router) handleRecordQuizAnswer(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	playerID := req.FormValue("player_id")
	reviewID := req.FormValue("review_id")
	answerID := req.FormValue("answer_id")

	if playerID == "" || reviewID == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "player_id and review_id are required",
		})
		return
	}

	// Get review to know correct answer
	review, err := r.reviewRepo.Get(reviewID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Review not found",
		})
		return
	}

	isCorrect := answerID == review.MovieID

	// Record answer
	answer := &domain.QuizAnswer{
		ID:             generateID(),
		PlayerID:       playerID,
		ReviewID:       reviewID,
		CorrectMovieID: review.MovieID,
		PlayerAnswerID: nil,
		IsCorrect:      isCorrect,
	}

	if answerID != "" {
		answer.PlayerAnswerID = &answerID
	}

	err = r.quizRepo.Create(answer)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to record answer",
		})
		return
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"correct":          isCorrect,
			"correct_movie_id": review.MovieID,
		},
	})
}

// handleGetPlayerStats returns player's quiz statistics
// GET /api/quiz/stats?player_id=xxx
func (r *Router) handleGetPlayerStats(w http.ResponseWriter, req *http.Request) {
	playerID := req.URL.Query().Get("player_id")
	if playerID == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "player_id is required",
		})
		return
	}

	total, correct, err := r.quizRepo.GetPlayerScores(playerID)
	if err != nil {
		jsonResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Player not found or no answers recorded",
		})
		return
	}

	accuracy := 0.0
	if total > 0 {
		accuracy = float64(correct) / float64(total) * 100
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"player_id": playerID,
			"total":     total,
			"correct":   correct,
			"accuracy":  accuracy,
		},
	})
}

// ===== REVIEWER ENDPOINTS =====

// handleListReviewers lists all reviewers
// GET /api/reviewers
func (r *Router) handleListReviewers(w http.ResponseWriter, req *http.Request) {
	reviewers, err := r.reviewerRepo.List()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to list reviewers",
		})
		return
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    reviewers,
	})
}

// handleGetReviewerByUsername gets a reviewer by username
// GET /api/reviewers/{username}
func (r *Router) handleGetReviewerByUsername(w http.ResponseWriter, req *http.Request) {
	username := req.PathValue("username")
	if username == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "username is required",
		})
		return
	}

	reviewers, err := r.reviewerRepo.List()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to fetch reviewers",
		})
		return
	}

	var found *domain.LetterboxdReviewer
	for _, reviewer := range reviewers {
		if reviewer.LetterboxdUsername == username {
			found = reviewer
			break
		}
	}

	if found == nil {
		jsonResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "Reviewer not found",
		})
		return
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    found,
	})
}

// ===== HELPERS =====

// generateID creates a simple unique ID (UUID-like)
func generateID() string {
	// In production, use github.com/google/uuid
	// For now, use a simple implementation
	rand := make([]byte, 16)
	for i := 0; i < len(rand); i++ {
		rand[i] = byte((i * 7) % 256)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		rand[0:4], rand[4:6], rand[6:8], rand[8:10], rand[10:16])
}
