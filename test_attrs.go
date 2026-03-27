package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0"),
	)

	c.OnHTML("article.production-viewing", func(e *colly.HTMLElement) {
		// Affiche les attributs disponibles
		title := e.ChildText("h2.primaryname a")
		href := e.ChildAttr("h2.primaryname a", "href")
		dataItemSlug := e.Attr("data-item-slug")
		dataObject := e.Attr("data-object-name")
		dataOwner := e.Attr("data-owner")

		fmt.Printf("Film: %s\n", strings.TrimSpace(title))
		fmt.Printf("  href: %s\n", href)
		fmt.Printf("  data-item-slug: %s\n", dataItemSlug)
		fmt.Printf("  data-object-name: %s\n", dataObject)
		fmt.Printf("  data-owner: %s\n", dataOwner)
		fmt.Println()
	})

	url := "https://letterboxd.com/66sceptre/reviews/"
	fmt.Printf("Checking attributes for %s\n\n", url)
	c.Visit(url)
}
