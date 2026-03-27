package httpapi

import (
	"net/http"

	"github.com/charmbracelet/log"
)

// Router gère les routes HTTP
type Router struct {
	handler *Handler
}

// NewRouter crée un nouveau router
func NewRouter(logger *log.Logger) *Router {
	return &Router{
		handler: NewHandler(logger),
	}
}

// Register enregistre toutes les routes
func (r *Router) Register() http.Handler {
	mux := http.NewServeMux()

	// Routes API
	mux.HandleFunc("GET /api/reviews", r.handler.FetchReviews)
	mux.HandleFunc("POST /api/game/start", r.handler.StartGame)
	mux.HandleFunc("GET /api/game/question", r.handler.GetCurrentQuestion)
	mux.HandleFunc("POST /api/game/answer", r.handler.SubmitAnswer)
	mux.HandleFunc("GET /api/game/score", r.handler.GetScore)
	mux.HandleFunc("GET /api/game/results", r.handler.GetResults)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, Response{
			Success: true,
			Data:    "Review Guess API v1.0",
		})
	})

	return mux
}
