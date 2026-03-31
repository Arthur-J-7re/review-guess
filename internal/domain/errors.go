package domain

import "fmt"

// Erreurs métier
var (
	ErrInvalidUsername = fmt.Errorf("invalid username")
)

// ScrapperError wraps scrapping errors
type ScrapperError struct {
	Username string
	Err      error
}

func (e *ScrapperError) Error() string {
	return fmt.Sprintf("scrapper error for %s: %v", e.Username, e.Err)
}
