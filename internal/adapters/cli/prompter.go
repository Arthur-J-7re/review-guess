package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Prompter gère les entrées utilisateur
type Prompter struct {
	reader *bufio.Reader
}

// NewPrompter crée un nouveau prompter
func NewPrompter() *Prompter {
	return &Prompter{
		reader: bufio.NewReader(os.Stdin),
	}
}

// PromptUsernames demande les pseudos Letterboxd
func (p *Prompter) PromptUsernames() ([]string, error) {
	PrintQuestion("Enter Letterboxd usernames (comma-separated)")
	fmt.Printf("%s> %s", ColorCyan, ColorReset)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("no usernames provided")
	}

	// Split et trim
	usernames := strings.Split(input, ",")
	for i, u := range usernames {
		usernames[i] = strings.TrimSpace(u)
	}

	return usernames, nil
}

// PromptQuestionCount demande le nombre de questions
func (p *Prompter) PromptQuestionCount() (int, error) {
	PrintQuestion("How many questions? (default 10)")
	fmt.Printf("%s> %s", ColorCyan, ColorReset)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return 10, nil
	}

	var count int
	_, err = fmt.Sscanf(input, "%d", &count)
	if err != nil || count <= 0 {
		return 0, fmt.Errorf("invalid number")
	}

	return count, nil
}

// PromptUsername demande un username (pour répondre)
func (p *Prompter) PromptUsername() (string, error) {
	fmt.Printf("%s> Username: %s", ColorCyan, ColorReset)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// PromptFilm demande un titre/slug de film (pour répondre)
func (p *Prompter) PromptFilm() (string, error) {
	fmt.Printf("%s> Film (title or slug): %s", ColorCyan, ColorReset)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(input), nil
}

// PromptContinue demande une touche pour continuer
func (p *Prompter) PromptContinue(text string) error {
	fmt.Printf("\n%s%s[Press Enter to continue]%s\n", ColorBold, ColorYellow, ColorReset)

	_, err := p.reader.ReadString('\n')
	return err
}

// PromptRetry demande si rejouer
func (p *Prompter) PromptRetry() (bool, error) {
	PrintQuestion("Play again? (y/n)")
	fmt.Printf("%s> %s", ColorCyan, ColorReset)

	input, err := p.reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	return strings.ToLower(strings.TrimSpace(input)) == "y", nil
}
