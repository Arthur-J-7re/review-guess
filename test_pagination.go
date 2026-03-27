package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowedDomains("letterboxd.com"),
	)

	c.OnHTML("a.next", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		text := e.Text()
		fmt.Printf("Next link found:\n  href: %s\n  text: %s\n", href, strings.TrimSpace(text))
	})

	c.OnHTML("a.paginate-page", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		text := strings.TrimSpace(e.Text())
		fmt.Printf("Page link: %s -> %s\n", text, href)
	})

	url := "https://letterboxd.com/66sceptre/reviews/"
	fmt.Printf("Checking pagination on %s\n\n", url)
	if err := c.Visit(url); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
