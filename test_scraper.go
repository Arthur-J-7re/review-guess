package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Test direct du scraper
	count := 0
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0"),
	)

	c.OnHTML("article.production-viewing", func(e *colly.HTMLElement) {
		count++
		title := e.ChildText("h2.primaryname a")
		slug := e.Attr("data-item-slug")
		fmt.Printf("%d. %s (slug: %s)\n", count, strings.TrimSpace(title), slug)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Printf("Visiting %s\n", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Printf("Error: %v\n", err)
	})

	url := "https://letterboxd.com/66sceptre/reviews/"
	fmt.Printf("Fetching %s...\n", url)
	if err := c.Visit(url); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		return
	}

	time.Sleep(1 * time.Second)
	fmt.Printf("\nTotal: %d reviews found\n", count)
}
