package httpapi

import (
	"net/http"
	"strings"

	"github.com/charmbracelet/log"

	"review-guess/internal/application"
	"review-guess/internal/infrastructure/scrapper"
)

// Router gère les routes HTTP
type Router struct {
	reviewService *application.ReviewService
	logger        *log.Logger
}

// NewRouter crée un nouveau router
func NewRouter(logger *log.Logger) *Router {
	scraper := scrapper.NewScrapper()
	reviewService := application.NewReviewService(scraper)

	return &Router{
		reviewService: reviewService,
		logger:        logger,
	}
}

// Register enregistre toutes les routes
func (r *Router) Register() http.Handler {
	mux := http.NewServeMux()

	// Routes API
	mux.HandleFunc("GET /api/reviews", r.handleGetReviews)
	mux.HandleFunc("GET /api/reviews/batch", r.handleGetReviewsBatch)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		jsonResponse(w, http.StatusOK, Response{
			Success: true,
			Data:    "Review Guess API v1.0",
		})
	})

	return mux
}

// handleGetReviews traite la requête GET /api/reviews
// Accepte les usernames en tant que query parameters répétés ou séparés par des virgules
func (r *Router) handleGetReviews(w http.ResponseWriter, req *http.Request) {
	usernamesParam := req.URL.Query()["username"]
	if len(usernamesParam) == 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "username parameter is required",
		})
		return
	}

	// Traiter les usernames en combinant les query params et en splitant par virgule
	var usernames []string
	for _, param := range usernamesParam {
		for _, username := range strings.Split(param, ",") {
			if trimmed := strings.TrimSpace(username); trimmed != "" {
				usernames = append(usernames, trimmed)
			}
		}
	}

	if len(usernames) == 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "at least one valid username is required",
		})
		return
	}

	reviews, err := r.reviewService.GetReviews(usernames...)
	if err != nil {
		r.logger.Error("Error fetching reviews", "error", err)
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

// handleGetReviewsBatch traite la requête GET /api/reviews/batch avec usernames séparés par des virgules
func (r *Router) handleGetReviewsBatch(w http.ResponseWriter, req *http.Request) {
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

	reviews, err := r.reviewService.GetReviews(usernames...)
	if err != nil {
		r.logger.Error("Error fetching reviews", "error", err)
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
