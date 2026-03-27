package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"review-guess/internal/application"
	"review-guess/internal/domain"
	"review-guess/internal/infrastructure/scrapper"

	"github.com/charmbracelet/log"
)

// Handler gère les requêtes HTTP
type Handler struct {
	gameService *application.GameService
	scraper     *scrapper.Scrapper
	logger      *log.Logger
}

// NewHandler crée un nouveau handler
func NewHandler(logger *log.Logger) *Handler {
	s := scrapper.NewScrapper()
	return &Handler{
		gameService: application.NewGameService(s),
		scraper:     s,
		logger:      logger,
	}
}

// Response JSON générique
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ReviewResponse pour une review unique
type ReviewResponse struct {
	Author   string `json:"author"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Content  string `json:"content"`
	Rating   int    `json:"rating"`
	Liked    bool   `json:"liked"`
	Spoilers bool   `json:"spoilers"`
}

// QuestionResponse pour une question du jeu
type QuestionResponse struct {
	Index      int            `json:"index"`
	Total      int            `json:"total"`
	Review     ReviewResponse `json:"review"`
	Difficulty float32        `json:"difficulty"`
}

// AnswerRequest pour soumettre une réponse
type AnswerRequest struct {
	GuessedAuthor string `json:"guessed_author"`
	GuessedFilm   string `json:"guessed_film"`
}

// AnswerResponse après avoir soumis une réponse
type AnswerResponse struct {
	Correct       bool   `json:"correct"`
	PartialAuthor bool   `json:"partial_author"`
	PartialFilm   bool   `json:"partial_film"`
	CorrectAuthor string `json:"correct_author"`
	CorrectFilm   string `json:"correct_film"`
	CorrectSlug   string `json:"correct_slug"`
	Points        int    `json:"points"`
	CurrentScore  int    `json:"current_score"`
	IsGameOver    bool   `json:"is_game_over"`
}

// FetchReviews récupère les reviews d'un utilisateur
func (h *Handler) FetchReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "username query parameter is required",
		})
		return
	}

	reviews, err := h.scraper.FetchUserReviews(username)
	if err != nil {
		h.logger.Error("Error fetching reviews", "err", err)
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to fetch reviews: " + err.Error(),
		})
		return
	}

	if len(reviews) == 0 {
		jsonResponse(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "No reviews found for this user",
		})
		return
	}

	reviewResponses := make([]ReviewResponse, len(reviews))
	for i, review := range reviews {
		reviewResponses[i] = ReviewResponse{
			Author:   review.Author,
			Title:    review.Film.Title,
			Slug:     review.Film.Slug,
			Content:  review.Content,
			Rating:   review.Rating,
			Liked:    review.Liked,
			Spoilers: review.Spoilers,
		}
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"count":   len(reviewResponses),
			"reviews": reviewResponses,
		},
	})
}

// StartGame initialise une nouvelle partie
func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req struct {
		Usernames     []string `json:"usernames"`
		QuestionCount int      `json:"question_count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if len(req.Usernames) == 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "At least one username is required",
		})
		return
	}

	if req.QuestionCount <= 0 {
		req.QuestionCount = 10
	}

	// Crée un nouveau GameService pour une nouvelle partie
	s := scrapper.NewScrapper()
	h.gameService = application.NewGameService(s)

	if err := h.gameService.LoadGame(req.Usernames, req.QuestionCount); err != nil {
		h.logger.Error("Error loading game", "err", err)
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to load game: " + err.Error(),
		})
		return
	}

	question, err := h.gameService.GetCurrentQuestion()
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "Failed to get first question: " + err.Error(),
		})
		return
	}

	results := h.gameService.GetResults()
	totalQuestions := results.TotalPoints / 100

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"message":          "Game started",
			"total_questions":  totalQuestions,
			"current_question": h.formatQuestion(question, 0, int(totalQuestions)),
		},
	})
}

// GetCurrentQuestion retourne la question actuelle
func (h *Handler) GetCurrentQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	question, err := h.gameService.GetCurrentQuestion()
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Game not started or no current question: " + err.Error(),
		})
		return
	}

	results := h.gameService.GetResults()
	index := len(results.Answers)
	total := results.TotalPoints / 100

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    h.formatQuestion(question, index, int(total)),
	})
}

// SubmitAnswer soumet une réponse
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var req AnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.GuessedAuthor) == "" || strings.TrimSpace(req.GuessedFilm) == "" {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "guessed_author and guessed_film are required",
		})
		return
	}

	answer, err := h.gameService.SubmitAnswer(req.GuessedAuthor, req.GuessedFilm)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	results := h.gameService.GetResults()

	// Get the question to display the correct answer
	var question *domain.Question
	if len(results.Answers) > 0 {
		lastAnswer := results.Answers[len(results.Answers)-1]
		// Find the question for this answer
		if lastAnswer.QuestionIdx < len(results.Answers) {
			// The question was for the previous index
		}
	}

	resp := AnswerResponse{
		Correct:       answer.IsCorrectUser && answer.IsCorrectFilm,
		PartialAuthor: answer.IsCorrectUser && !answer.IsCorrectFilm,
		PartialFilm:   answer.IsCorrectFilm && !answer.IsCorrectUser,
		CorrectAuthor: answer.GuessedUser,
		CorrectFilm:   answer.GuessedFilm,
		CorrectSlug:   "",
		Points:        answer.Points,
		CurrentScore:  results.TotalScore,
		IsGameOver:    h.gameService.IsGameOver(),
	}

	// If we have a question, fill in the correct answer
	if question != nil {
		resp.CorrectAuthor = question.Review.Author
		resp.CorrectFilm = question.Review.Film.Title
		resp.CorrectSlug = question.Review.Film.Slug
	}

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data:    resp,
	})
}

// GetScore retourne le score actuel
func (h *Handler) GetScore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	score := h.gameService.GetScore()
	results := h.gameService.GetResults()
	total := results.TotalPoints / 100

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"current_score":   score,
			"answered":        len(results.Answers),
			"total_questions": total,
			"is_game_over":    h.gameService.IsGameOver(),
		},
	})
}

// GetResults retourne les résultats finaux
func (h *Handler) GetResults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonResponse(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	if !h.gameService.IsGameOver() {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Game is not over yet",
		})
		return
	}

	results := h.gameService.GetResults()
	grade := getGrade(int(results.Percentage))

	jsonResponse(w, http.StatusOK, Response{
		Success: true,
		Data: map[string]interface{}{
			"score":        results.TotalScore,
			"total_points": results.TotalPoints,
			"percentage":   int(results.Percentage),
			"grade":        grade,
			"answered":     len(results.Answers),
			"total":        results.TotalPoints / 100,
		},
	})
}

// Helper functions

func (h *Handler) formatQuestion(question *domain.Question, index, total int) QuestionResponse {
	if question == nil {
		return QuestionResponse{}
	}

	return QuestionResponse{
		Index: index + 1,
		Total: total,
		Review: ReviewResponse{
			Author:   question.Review.Author,
			Title:    question.Review.Film.Title,
			Slug:     question.Review.Film.Slug,
			Content:  question.Review.Content,
			Rating:   question.Review.Rating,
			Liked:    question.Review.Liked,
			Spoilers: question.Review.Spoilers,
		},
		Difficulty: question.Difficulty,
	}
}

func getGrade(percentage int) string {
	if percentage >= 90 {
		return "A+"
	}
	if percentage >= 80 {
		return "A"
	}
	if percentage >= 70 {
		return "B"
	}
	if percentage >= 60 {
		return "C"
	}
	return "F"
}

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
