package cli

import (
	"fmt"
	"strings"
)

// Color codes pour le terminal
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

// ClearScreen efface l'écran du terminal
func ClearScreen() {
	fmt.Print("\033[2J\033[H")
}

// PrintHeader affiche un titre stylisé
func PrintHeader(text string) {
	fmt.Println()
	fmt.Printf("%s%s=== %s ===%s\n", ColorBold, ColorBlue, text, ColorReset)
	fmt.Println()
}

// PrintSuccess affiche un message succès en vert
func PrintSuccess(text string) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, text, ColorReset)
}

// PrintError affiche un message erreur en rouge
func PrintError(text string) {
	fmt.Printf("%s✗ %s%s\n", ColorRed, text, ColorReset)
}

// PrintWarning affiche un message warning en jaune
func PrintWarning(text string) {
	fmt.Printf("%s⚠ %s%s\n", ColorYellow, text, ColorReset)
}

// PrintInfo affiche un message info en cyan
func PrintInfo(text string) {
	fmt.Printf("%s→ %s%s\n", ColorCyan, text, ColorReset)
}

// PrintQuestion affiche une question stylisée
func PrintQuestion(question string) {
	fmt.Printf("\n%s%s❓ %s%s\n", ColorBold, ColorPurple, question, ColorReset)
}

// PrintReviewBox affiche une review dans une belle boîte
func PrintReviewBox(content string) {
	width := 70

	fmt.Println()
	fmt.Printf("%s%s┌", ColorBold, ColorCyan)
	for i := 0; i < width-2; i++ {
		fmt.Print("─")
	}
	fmt.Printf("┐%s\n", ColorReset)

	// Wrap text
	lines := wrapText(content, width-4)
	for _, line := range lines {
		fmt.Printf("%s%s│ %s%-*s │%s\n", ColorBold, ColorCyan, ColorReset, width-4, line, ColorReset)
	}

	fmt.Printf("%s%s└", ColorBold, ColorCyan)
	for i := 0; i < width-2; i++ {
		fmt.Print("─")
	}
	fmt.Printf("┘%s\n", ColorReset)
	fmt.Println()
}

// wrapText wraps text to a given width
func wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	var currentLine string

	for _, word := range words {
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
		}
		if currentLine == "" {
			currentLine = word
		} else {
			currentLine += " " + word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// PrintAnswer affiche le résultat d'une réponse
func PrintAnswer(correct bool, author string, film string) {
	if correct {
		PrintSuccess("Correct answer!")
	} else {
		PrintError("Wrong answer!")
	}
	fmt.Printf("  By: %s | Film: %s\n", author, film)
}

// PrintScore affiche le score avec une barre
func PrintScore(current, total int) {
	percentage := float32(current) / float32(total) * 100
	barWidth := 30
	filledWidth := int(percentage / 100 * float32(barWidth))

	fmt.Printf("\nScore: %s%d/%d%s (", ColorBold, current, total, ColorReset)

	// Barre de progression
	if percentage >= 75 {
		fmt.Printf("%s", ColorGreen)
	} else if percentage >= 50 {
		fmt.Printf("%s", ColorYellow)
	} else {
		fmt.Printf("%s", ColorRed)
	}

	for i := 0; i < filledWidth; i++ {
		fmt.Print("█")
	}
	for i := filledWidth; i < barWidth; i++ {
		fmt.Print("░")
	}

	fmt.Printf("%s) %.1f%%\n", ColorReset, percentage)
}

// PrintResults affiche l'écran final des résultats
func PrintResults(score, total int, answers int) {
	ClearScreen()
	PrintHeader("Game Over")

	percentage := float32(score) / float32(total) * 100

	// Grade
	var grade, gradeStyle string
	if percentage >= 90 {
		grade = "A+"
		gradeStyle = ColorGreen
	} else if percentage >= 80 {
		grade = "A"
		gradeStyle = ColorGreen
	} else if percentage >= 70 {
		grade = "B"
		gradeStyle = ColorCyan
	} else if percentage >= 60 {
		grade = "C"
		gradeStyle = ColorYellow
	} else {
		grade = "F"
		gradeStyle = ColorRed
	}

	fmt.Printf("%s%s%s%s %s%.1f%%%s\n\n", ColorBold, gradeStyle, grade, ColorReset, ColorBold, percentage, ColorReset)

	PrintScore(score, total)
	fmt.Printf("Questions answered: %d\n\n", answers)

	// Messages motivants
	if percentage == 100 {
		PrintSuccess("Perfect score! You're a movie master!")
	} else if percentage >= 80 {
		PrintSuccess("Excellent! You really know your movies!")
	} else if percentage >= 60 {
		PrintInfo("Nice job! Keep watching more movies!")
	} else {
		PrintWarning("Need to watch more movies? Try again!")
	}

	fmt.Println()
}

// PrintProgressBar affiche une barre de progression multi-lignes
func PrintProgressBar(current, total int) {
	fmt.Printf("%s[%d/%d]%s ", ColorCyan, current, total, ColorReset)

	barWidth := 20
	filledWidth := int(float32(current) / float32(total) * float32(barWidth))

	for i := 0; i < filledWidth; i++ {
		fmt.Print("█")
	}
	for i := filledWidth; i < barWidth; i++ {
		fmt.Print("░")
	}

	percentage := float32(current) / float32(total) * 100
	fmt.Printf(" %.0f%%\n", percentage)
}
