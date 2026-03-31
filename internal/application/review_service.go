package application

import (
	"fmt"
	"strings"

	"review-guess/internal/domain"
)

type ReviewService struct {
	provider domain.ReviewProvider
}

func NewReviewService(provider domain.ReviewProvider) *ReviewService {
	return &ReviewService{
		provider: provider,
	}
}

// GetReviews récupère les reviews des utilisateurs spécifiés
func (s *ReviewService) GetReviews(usernames ...string) (*domain.Reviews, error) {
	if len(usernames) == 0 {
		return nil, fmt.Errorf("at least one username is required")
	}

	var allReviews []*domain.Review

	for _, username := range usernames {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}

		reviews, err := s.provider.FetchUserReviews(username)
		if err != nil {
			return nil, err
		}

		allReviews = append(allReviews, reviews...)
	}

	return &domain.Reviews{
		Count:   len(allReviews),
		Reviews: allReviews,
	}, nil
}
