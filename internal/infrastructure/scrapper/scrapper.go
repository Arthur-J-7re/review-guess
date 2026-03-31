package scrapper

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"review-guess/internal/domain"
	"review-guess/internal/infrastructure/cache"

	"github.com/charmbracelet/log"
	"github.com/gocolly/colly/v2"
)

const (
	maxConcurrentPages = 15
	totalTimeout       = 60 * time.Second
	pageDelay          = 3 * time.Second
	randomDelay        = 2 * time.Second
	reviewsCacheTTL    = 10 * time.Minute
)

var reviewsCache = cache.New()

// Scrapper récupère les reviews de Letterboxd
type Scrapper struct {
	baseURL          string
	logger           *log.Logger
	NewCollector     CollectorFactory
	GetTotalPages    func(url string) (int, error)
	GetReviewsOnPage func(url string, page int) ([]*domain.Review, error)
}

type CollectorFactory func() *colly.Collector

// NewScrapper crée un nouveau scrapper
func NewScrapper() *Scrapper {
	scrapper := &Scrapper{
		baseURL: "https://letterboxd.com",
		logger:  log.New(os.Stderr),
	}

	scrapper.NewCollector = func() *colly.Collector { return colly.NewCollector() }
	scrapper.GetTotalPages = func(url string) (int, error) {
		return scrapper.getTotalPagesImpl(url)
	}
	scrapper.GetReviewsOnPage = func(url string, page int) ([]*domain.Review, error) {
		return scrapper.getReviewsOnPageImpl(url, page)
	}

	return scrapper
}

// FetchUserReviews récupère toutes les reviews d'un utilisateur
func (s *Scrapper) FetchUserReviews(username string) ([]*domain.Review, error) {
	if strings.TrimSpace(username) == "" {
		return nil, domain.ErrInvalidUsername
	}

	reviewsURL := fmt.Sprintf("%s/%s/reviews/", s.baseURL, strings.ToLower(username))

	// Check cache first
	if cached, found := reviewsCache.Get(reviewsURL); found {
		reviews := cached.([]*domain.Review)
		s.logger.Info("Reviews cache hit", "username", username, "count", len(reviews))
		return reviews, nil
	}

	s.logger.Info("Reviews cache miss, scraping", "username", username)
	s.logger.Info("Getting total pages", "username", username)
	totalPages, err := s.GetTotalPages(reviewsURL)
	if err != nil {
		return nil, &domain.ScrapperError{Username: username, Err: err}
	}

	if totalPages == 0 {
		s.logger.Warn("No pages found", "username", username)
		return []*domain.Review{}, nil
	}

	s.logger.Info("Fetching reviews", "username", username, "total_pages", totalPages)

	ctx, cancel := context.WithTimeout(context.Background(), totalTimeout)
	defer cancel()

	estimatedReviews := totalPages * 20
	allReviews := make([]*domain.Review, 0, estimatedReviews)

	type pageResult struct {
		reviews []*domain.Review
		page    int
		err     error
	}

	resultCh := make(chan pageResult, totalPages)
	semaphore := make(chan struct{}, maxConcurrentPages)

	var wg sync.WaitGroup

	for page := 1; page <= totalPages; page++ {
		select {
		case <-ctx.Done():
			s.logger.Warn("Context timeout before fetching all pages")
			break
		default:
		}

		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				resultCh <- pageResult{err: ctx.Err(), page: pageNum}
				return
			}

			reviews, err := s.GetReviewsOnPage(reviewsURL, pageNum)
			resultCh <- pageResult{
				reviews: reviews,
				page:    pageNum,
				err:     err,
			}
		}(page)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	errors := make([]error, 0)
	successCount := 0

	for result := range resultCh {
		if result.err != nil {
			errors = append(errors, fmt.Errorf("page %d: %w", result.page, result.err))
			s.logger.Warn("Failed to fetch page", "page", result.page, "err", result.err)
			continue
		}

		allReviews = append(allReviews, result.reviews...)
		successCount++
		s.logger.Debug("Fetched page", "page", result.page, "reviews", len(result.reviews))
	}

	if len(errors) > 0 {
		if successCount == 0 {
			return nil, fmt.Errorf("all pages failed: %w", errors[0])
		}

		if len(errors) > totalPages/2 {
			return nil, fmt.Errorf("too many failures (%d/%d): %w",
				len(errors), totalPages, errors[0])
		}

		s.logger.Warn("Partial success fetching reviews",
			"username", username,
			"success", fmt.Sprintf("%d/%d", successCount, totalPages),
			"errors", len(errors))
	}

	s.logger.Info("Fetched all reviews", "username", username, "count", len(allReviews), "pages", totalPages)

	// Cache the reviews
	if len(allReviews) > 0 {
		reviewsCache.Set(reviewsURL, allReviews, reviewsCacheTTL)
		s.logger.Info("Cached reviews", "username", username, "ttl", "10min")
	}

	return allReviews, nil
}

// getTotalPagesImpl récupère le nombre total de pages de reviews
func (s *Scrapper) getTotalPagesImpl(reviewsURL string) (int, error) {
	totalPages := 1

	collector := s.NewCollector()
	collector.OnHTML("div.paginate-pages ul", func(e *colly.HTMLElement) {
		e.ForEach("li.paginate-page a", func(_ int, el *colly.HTMLElement) {
			if n, err := strconv.Atoi(el.Text); err == nil && n > totalPages {
				totalPages = n
			}
		})
	})

	if err := collector.Visit(reviewsURL); err != nil {
		return 0, err
	}

	return totalPages, nil
}

// getReviewsOnPageImpl récupère les reviews d'une page spécifique
func (s *Scrapper) getReviewsOnPageImpl(reviewsURL string, page int) ([]*domain.Review, error) {
	pageURL := fmt.Sprintf("%s/films/page/%d", reviewsURL, page)
	var reviews []*domain.Review

	collector := s.NewCollector()

	// Rate limiting avec délai aléatoire
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "letterboxd.com",
		Delay:       pageDelay,
		RandomDelay: randomDelay,
	})

	collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Referer", "https://letterboxd.com")
	})

	collector.OnHTML("article.production-viewing", func(e *colly.HTMLElement) {
		review := s.parseReviewElement(e)
		if review != nil {
			reviews = append(reviews, review)
		}
	})

	if err := collector.Visit(pageURL); err != nil {
		return nil, err
	}

	return reviews, nil
}

// parseReviewElement parse un élément review HTML
func (s *Scrapper) parseReviewElement(e *colly.HTMLElement) *domain.Review {
	// Vérifier que c'est bien une review (pas un like/bookmark)
	objectName := e.Attr("data-object-name")
	if strings.TrimSpace(objectName) != "review" {
		return nil
	}

	// Récupère le username/author depuis data-owner
	author := e.Attr("data-owner")
	if author == "" {
		return nil
	}

	// Récupère le film slug depuis le href du lien
	filmLink := e.ChildAttr("h2.primaryname a", "href")
	if filmLink == "" {
		return nil
	}

	// Extrait le slug depuis l'URL /..../film/alter-ego-2026/ -> alter-ego-2026
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

	// Récupère le titre du film
	filmTitle := e.ChildText("h2.primaryname a")
	filmTitle = strings.TrimSpace(filmTitle)

	if filmTitle == "" {
		return nil
	}

	// Récupère le contenu de la review
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

	// Récupère spoiler marker
	spoilers := e.ChildAttr("span", "data-spoiler") != ""

	review := &domain.Review{
		Author:   author,
		Title:    filmTitle,
		Slug:     filmSlug,
		Content:  content,
		Rating:   rating,
		Liked:    liked,
		Spoilers: spoilers,
	}

	return review
}

// parseRating extrait la note depuis the aria-label (e.g., "★★★★★", "★★★★½", "★★")
// Garde les .5 pour les demi-étoiles (Letterboxd uses 0.5 increments)
func parseRating(ariaLabel string) float64 {
	ariaLabel = strings.TrimSpace(ariaLabel)

	fullStars := strings.Count(ariaLabel, "★")
	halfStars := strings.Count(ariaLabel, "½")

	rating := float64(fullStars)
	if halfStars > 0 {
		rating += 0.5 // Add 0.5 for each half star
	}

	if rating > 5 {
		rating = 5
	}

	return rating
}
