package scrapper

import (
	"fmt"
	"os"
	"strings"
	"time"

	"review-guess/internal/domain"

	"github.com/charmbracelet/log"
	"github.com/gocolly/colly/v2"
)

// Scrapper récupère les reviews de Letterboxd
type Scrapper struct {
	baseURL string
	logger  *log.Logger
}

// NewScrapper crée un nouveau scrapper
func NewScrapper() *Scrapper {
	return &Scrapper{
		baseURL: "https://letterboxd.com",
		logger:  log.New(os.Stderr),
	}
}

// FetchUserReviews récupère toutes les reviews d'un utilisateur
func (s *Scrapper) FetchUserReviews(username string) ([]*domain.Review, error) {
	if strings.TrimSpace(username) == "" {
		return nil, domain.ErrInvalidUsername
	}

	var reviews []*domain.Review
	pageURL := fmt.Sprintf("%s/%s/reviews/", s.baseURL, strings.ToLower(username))

	// Crée un collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("letterboxd.com"),
	)

	// Add headers to look more like a real browser
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Referer", "https://letterboxd.com")
		s.logger.Info("Visiting", "url", r.URL)
	})

	// Rate limiting - be more conservative
	c.Limit(&colly.LimitRule{
		DomainGlob:  "letterboxd.com",
		Delay:       3 * time.Second,
		RandomDelay: 2 * time.Second,
	})

	// Parse chaque review dans la liste
	c.OnHTML("article.production-viewing", func(e *colly.HTMLElement) {
		review := s.parseReviewElement(e, username)
		if review != nil {
			reviews = append(reviews, review)
		}
	})

	// Pagination: s'il y a une page suivante
	var nextPageURL string
	c.OnHTML("a.next", func(e *colly.HTMLElement) {
		nextPageURL = e.Attr("href")
		s.logger.Info("Found next link", "href", nextPageURL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		s.logger.Error("Error scrapping", "username", username, "err", err)
	})

	// (OnRequest handler already defined above for headers)

	// Fetch première page
	if err := c.Visit(pageURL); err != nil {
		return nil, &domain.ScrapperError{Username: username, Err: err}
	}

	// Fetch pages suivantes (pagination)
	pageCount := 1
	for nextPageURL != "" && pageCount < 50 { // Limite à 50 pages (600+ reviews)
		s.logger.Info("Fetching next page", "page", pageCount+1, "url", nextPageURL)

		// Délai important entre les pages pour éviter le blocage
		time.Sleep(5 * time.Second)

		currentPageURL := nextPageURL
		nextPageURL = "" // Reset pour chercher le lien de la PROCHAINE page

		if err := c.Visit(s.baseURL + currentPageURL); err != nil {
			s.logger.Warn("Error on page", "page", pageCount+1, "url", currentPageURL, "err", err)
			// Continuer même en cas d'erreur pour les pages suivantes
			if strings.Contains(err.Error(), "Forbidden") {
				s.logger.Info("Got 403 Forbidden, trying again with longer delay", "page", pageCount+1)
				// Attendre plus longtemps et réessayer UNE fois
				time.Sleep(10 * time.Second)
				if err := c.Visit(s.baseURL + currentPageURL); err != nil {
					s.logger.Warn("Second attempt also failed", "page", pageCount+1, "err", err)
					break
				}
			} else {
				break
			}
		}
		pageCount++
	}

	s.logger.Info("Fetched reviews", "username", username, "count", len(reviews), "pages", pageCount)

	return reviews, nil
}

// parseReviewElement parse un élément review HTML
func (s *Scrapper) parseReviewElement(e *colly.HTMLElement, username string) *domain.Review {
	// Vérifier que c'est bien une review (pas un like/bookmark)
	objectName := e.Attr("data-object-name")
	if strings.TrimSpace(objectName) != "review" {
		return nil // Ce n'est pas une review
	}

	// Vérifier que c'est écrit par l'utilisateur (case-insensitive)
	owner := e.Attr("data-owner")
	if owner == "" || !strings.EqualFold(owner, username) {
		return nil // Ce n'est pas une review écrite par cet utilisateur
	}

	// Récupère le film slug depuis le href du lien
	filmLink := e.ChildAttr("h2.primaryname a", "href")
	if filmLink == "" {
		return nil
	}

	// Extrait le slug depuis l'URL /66sceptre/film/alter-ego-2026/ -> alter-ego-2026
	var filmSlug string
	parts := strings.Split(strings.TrimSpace(filmLink), "/")
	for i, part := range parts {
		if part == "film" && i+1 < len(parts) {
			filmSlug = parts[i+1]
			break
		}
	}

	if filmSlug == "" {
		return nil
	}

	// Récupère le titre du film depuis h2.primaryname a
	filmTitle := e.ChildText("h2.primaryname a")
	filmTitle = strings.TrimSpace(filmTitle)

	if filmTitle == "" {
		return nil
	}

	// Récupère le contenu de la review depuis div.js-review-body
	content := e.ChildText("div.js-review-body")
	content = strings.TrimSpace(content)

	// Si pas de contenu, on le saute
	if content == "" {
		return nil
	}

	// Récupère la note depuis aria-label de SVG .glyph.-rating
	ratingText := e.ChildAttr("svg.glyph.-rating", "aria-label")
	rating := parseRating(ratingText)

	// Récupère le like: cherche SVG avec class inline-liked
	liked := e.ChildAttr("svg.inline-liked", "aria-label") != ""

	// Récupère spoiler marker: cherche span avec "spoiler" ou data-spoiler
	spoilers := e.ChildAttr("span", "data-spoiler") != "" || strings.Contains(e.ChildText(""), "spoiler")

	review := &domain.Review{
		ID:     fmt.Sprintf("%s-%s", username, filmSlug),
		Author: username,
		Film: &domain.Film{
			Slug:  filmSlug,
			Title: filmTitle,
		},
		Content:  content,
		Rating:   rating,
		Liked:    liked,
		Spoilers: spoilers,
	}

	return review
}

// parseRating extrait la note desde the aria-label (e.g., "★★★★★", "★★★★½", "★★")
func parseRating(ariaLabel string) int {
	ariaLabel = strings.TrimSpace(ariaLabel)

	// Count full stars (★) and half stars (½)
	fullStars := strings.Count(ariaLabel, "★")
	halfStars := strings.Count(ariaLabel, "½")

	rating := fullStars
	if halfStars > 0 {
		// Si on a une demi-étoile, on arrondit à l'entier inférieur
		// Pour parser correctement, on garde juste le nombre d'étoiles pleines
		// car notre système de rating utilise des entiers
	}

	// Si aria-label est vide, pas de note = Watched
	if rating == 0 && ariaLabel == "" {
		return 0
	}

	return rating
}
