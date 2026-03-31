package scrapper

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HTTPFetcher récupère le contenu d'une URL en utilisant net/http (bypasse les restrictions de colly)
func HTTPFetcher(url string) ([]byte, error) {
	// Create HTTP client with cookie jar to persist cookies across requests
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow redirects
			return nil
		},
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body[:min(200, len(body))])
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, bodyStr)
	}

	return io.ReadAll(resp.Body)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParsePaginationFromHTML extracte le nombre total de pages du HTML
func ParsePaginationFromHTML(htmlBody string) (int, error) {
	// Cherche les liens <a> dans <div class="paginate-pages">
	// Regex: <li class="paginate-page"><a href=".+">(\d+)</a></li>
	re := regexp.MustCompile(`<li[^>]*paginate-page[^>]*><a[^>]*>(\d+)</a></li>`)
	matches := re.FindAllStringSubmatch(htmlBody, -1)

	maxPage := 1
	for _, match := range matches {
		if len(match) > 1 {
			if page, err := strconv.Atoi(match[1]); err == nil && page > maxPage {
				maxPage = page
			}
		}
	}

	return maxPage, nil
}

// unescapeHTML convert des HTML entities simples(&quot; &amp; etc)
func unescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	return s
}
