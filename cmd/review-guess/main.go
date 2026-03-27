package main

import (
	"fmt"
	"net/http"
	"os"

	"review-guess/internal/adapters/httpapi"

	"github.com/charmbracelet/log"
)

func main() {
	logger := log.New(os.Stderr)

	// Create router
	router := httpapi.NewRouter(logger)
	mux := router.Register()

	// Start server
	port := ":8080"
	logger.Info("Starting Review Guess API", "port", port)
	logger.Info("Health check: http://localhost:8080/health")
	logger.Info("API docs:")
	logger.Info("  GET  /api/reviews?username={username} - Fetch user reviews")
	logger.Info("  POST /api/game/start - Start a new game")
	logger.Info("  GET  /api/game/question - Get current question")
	logger.Info("  POST /api/game/answer - Submit an answer")
	logger.Info("  GET  /api/game/score - Get current score")
	logger.Info("  GET  /api/game/results - Get final results")
	fmt.Println()

	if err := http.ListenAndServe(port, mux); err != nil {
		logger.Fatal("Server error", "err", err)
	}
}
