package scrapper

import (
	"math/rand"
	"strings"

	"review-guess/internal/domain"
)

// FilterQualityReviews filtre les reviews pour garder seulement les bonnes
func FilterQualityReviews(reviews []*domain.Review) []*domain.Review {
	var filtered []*domain.Review

	for _, review := range reviews {
		// Skip les reviews trop courtes (< 30 chars)
		if len(strings.TrimSpace(review.Content)) < 30 {
			continue
		}

		// Skip les "Watched" sans contenu
		if strings.ToLower(review.Content) == "watched" {
			continue
		}

		// Garde si: contenu bon OU rating ≥ 3 OU liked
		if len(review.Content) > 30 || review.Rating >= 3 || review.Liked {
			filtered = append(filtered, review)
		}
	}

	return filtered
}

// ShuffleReviews mélange les reviews aléatoirement (Fisher-Yates)
func ShuffleReviews(reviews []*domain.Review) {
	for i := len(reviews) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		reviews[i], reviews[j] = reviews[j], reviews[i]
	}
}

// CalculateDifficulty calcule la difficulté d'une review
// Short = hard (peu d'indices), Long = easy (bcp d'indices)
func CalculateDifficulty(review *domain.Review) float32 {
	difficulty := float32(1.0)

	// Court = difficile
	contentLen := len(review.Content)
	if contentLen < 100 {
		difficulty *= 1.5
	}

	// Pas de rating = difficile (moins d'infos)
	if review.Rating == 0 {
		difficulty *= 1.3
	}

	// Spoilers = facile (review usuellement plus longue/détaillée)
	if review.Spoilers {
		difficulty *= 0.8
	}

	// Limiter entre 0.3 et 3.0
	if difficulty < 0.3 {
		difficulty = 0.3
	}
	if difficulty > 3.0 {
		difficulty = 3.0
	}

	return difficulty
}
