package main

import (
	"fmt"
	"review-guess/internal/infrastructure/scrapper"
)

func main() {
	s := scrapper.NewScrapper()
	reviews, err := s.FetchUserReviews("66sceptre")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Total reviews: %d\n\n", len(reviews))
	for i, r := range reviews {
		fmt.Printf("%d. %s\n", i+1, r.Film.Title)
		fmt.Printf("   Author: %s, Rating: %d, ID: %s\n", r.Author, r.Rating, r.ID)
	}
}
