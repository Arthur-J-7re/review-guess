package application

import (
	"fmt"
	"time"

	"review-guess/internal/domain"
	"review-guess/internal/infrastructure/scrapper"
)

// ReviewerUpdateService gère la mise à jour intelligente des reviewers
type ReviewerUpdateService struct {
	reviewerRepo       domain.LetterboxdReviewerRepository
	reviewRepo         domain.ReviewRepository
	movieRepo          domain.MovieRepository
	scraper            *scrapper.Scrapper
	cacheTokenDuration time.Duration // 1h par défaut
}

// NewReviewerUpdateService crée un nouveau service
func NewReviewerUpdateService(
	reviewerRepo domain.LetterboxdReviewerRepository,
	reviewRepo domain.ReviewRepository,
	movieRepo domain.MovieRepository,
	scraper *scrapper.Scrapper,
) *ReviewerUpdateService {
	return &ReviewerUpdateService{
		reviewerRepo:       reviewerRepo,
		reviewRepo:         reviewRepo,
		movieRepo:          movieRepo,
		scraper:            scraper,
		cacheTokenDuration: 1 * time.Hour,
	}
}

// UpdateReviewerData met à jour les données d'un reviewer
// Utilise le cache token pour éviter les requêtes répétées (1h)
// Ne scrape que les pages nécessaires
func (s *ReviewerUpdateService) UpdateReviewerData(username string) (*domain.LetterboxdReviewer, error) {
	// 1. Chercher si le reviewer existe en base
	reviewer, err := s.reviewerRepo.GetByUsername(username)
	if err != nil {
		// Reviewer n'existe pas, on va le créer après le scraping
		reviewer = nil
	}

	// 2. Vérifier le cache token (1h)
	if reviewer != nil && reviewer.LastScrappedAt != nil {
		if time.Since(*reviewer.LastScrappedAt) < s.cacheTokenDuration {
			// Encore en cache, retourner sans scraper
			return reviewer, nil
		}
	}

	// 3. Scraper les nouvelles données
	reviews, err := s.scraper.FetchUserReviews(username)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape reviews: %w", err)
	}

	now := time.Now()

	// 4. Créer ou mettre à jour le reviewer
	if reviewer == nil {
		// Créer un nouveau reviewer
		reviewer = &domain.LetterboxdReviewer{
			ID:                     generateID(),
			LetterboxdUsername:     username,
			LastReviewPageScrapped: 1,
			LastMoviePageScrapped:  1,
			LastScrappedAt:         &now,
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		err = s.reviewerRepo.Create(reviewer)
		if err != nil {
			return nil, fmt.Errorf("failed to create reviewer: %w", err)
		}
	} else {
		// Update le reviewer avec le nouveau timestamp
		reviewer.LastScrappedAt = &now
		reviewer.UpdatedAt = now
		err = s.reviewerRepo.Update(reviewer)
		if err != nil {
			return nil, fmt.Errorf("failed to update reviewer: %w", err)
		}
	}

	// 5. Ajouter les reviews et films en base
	err = s.storeReviewsAndMovies(reviewer.ID, reviews)
	if err != nil {
		return nil, fmt.Errorf("failed to store reviews and movies: %w", err)
	}

	return reviewer, nil
}

// storeReviewsAndMovies sauvegarde les reviews et crée les films s'ils sont nouveaux
func (s *ReviewerUpdateService) storeReviewsAndMovies(reviewerID string, reviews []*domain.Review) error {
	for _, review := range reviews {
		// 1. Vérifier si le film existe par son slug Letterboxd
		movie, err := s.movieRepo.GetByLetterboxdSlug(review.MovieID)
		if err != nil || movie == nil {
			// Nouveau film détecté - créer un stub
			movie = &domain.Movie{
				ID:             generateID(),
				LetterboxdSlug: review.MovieID,
				Title:          "Unknown",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			err = s.movieRepo.Create(movie)
			if err != nil {
				return fmt.Errorf("failed to create movie: %w", err)
			}
		}

		// 2. Vérifier si la review existe déjà
		existingReview, err := s.reviewRepo.GetByReviewerAndMovie(reviewerID, movie.ID)
		if err == nil && existingReview != nil {
			// Review existe déjà, skip pour éviter les doublons
			continue
		}

		// 3. Ajouter la nouvelle review
		review.ID = generateID()
		review.ReviewerID = reviewerID
		review.MovieID = movie.ID
		review.Usable = true // Par défaut, les reviews sont utilisables pour le quiz
		review.CreatedAt = time.Now()
		review.UpdatedAt = time.Now()

		err = s.reviewRepo.Create(review)
		if err != nil {
			return fmt.Errorf("failed to create review: %w", err)
		}
	}

	return nil
}

// generateID génère un ID simple basé sur le timestamp
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
