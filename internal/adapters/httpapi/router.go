package httpapi

import (
	"net/http"

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
func (r *Router) handleGetReviews(w http.ResponseWriter, req *http.Request) {
	usernames := req.URL.Query()["username"]
	if len(usernames) == 0 {
		jsonResponse(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "username parameter is required",
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
