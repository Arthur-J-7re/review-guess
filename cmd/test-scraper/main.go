package main

import (
	"fmt"
	"log"
	"os"

	"review-guess/internal/infrastructure/scrapper"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/test-scraper/main.go <username>")
		fmt.Println("Example: go run cmd/test-scraper/main.go theprincessbride")
		return
	}

	username := os.Args[1]
	fmt.Printf("Scraping reviews for: %s\n", username)

	s := scrapper.NewScrapper()
	reviews, err := s.FetchUserReviews(username)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("\n✓ Fetched %d reviews\n\n", len(reviews))

	// Affiche les premiers reviews
	for i, review := range reviews {
		if i >= 5 {
			fmt.Println("... (truncated)")
			break
		}
		fmt.Printf("[%d] %s - %s\n", i+1, review.Author, review.Film.Title)
		fmt.Printf("    Content: %s...\n", review.Content[:min(80, len(review.Content))])
		fmt.Printf("    Rating: %d, Liked: %v\n\n", review.Rating, review.Liked)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
