package domain

import "fmt"

// Erreurs métier
var (
	ErrInvalidUsername  = fmt.Errorf("invalid username")
	ErrNotEnoughReviews = fmt.Errorf("not enough reviews to generate questions")
	ErrGameNotStarted   = fmt.Errorf("game not started")
	ErrGameOver         = fmt.Errorf("game is over")
	ErrInvalidAnswer    = fmt.Errorf("invalid answer")
)

// ScrapperError wraps scrapping errors
type ScrapperError struct {
	Username string
	Err      error
}

func (e *ScrapperError) Error() string {
	return fmt.Sprintf("scrapper error for %s: %v", e.Username, e.Err)
}
