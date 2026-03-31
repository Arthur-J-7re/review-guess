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
	logger.Info("Available endpoints:")
	logger.Info("  GET  /health - Health check")
	logger.Info("  GET  /api/reviews?username={username} - Fetch user reviews")
	fmt.Println()

	if err := http.ListenAndServe(port, mux); err != nil {
		logger.Fatal("Server error", "err", err)
	}
}
