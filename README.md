# Review Guess API

A simple Go backend API for fetching Letterboxd reviews. The API allows retrieving reviews from one or multiple Letterboxd users.

## Prerequisites

- Go 1.24+
- Port 8080 available

## Installation & Running

```bash
# Build
go build ./cmd/review-guess

# Run
./review-guess
```

The API will start on `http://localhost:8080`

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Get Reviews
```bash
curl "http://localhost:8080/api/reviews?username=alice"
curl "http://localhost:8080/api/reviews?username=alice&username=bob"
```

You can specify multiple users by repeating the `username` parameter.

## Project Structure

```
cmd/review-guess/
  └── main.go              # Entry point

internal/
  ├── adapters/
  │   └── httpapi/         # HTTP handlers and router
  ├── application/
  │   └── review_service.go  # Business logic
  ├── domain/
  │   ├── models.go        # Domain models (Review, Reviews)
  │   ├── errors.go        # Error definitions
  │   └── ports.go         # Interfaces (ReviewProvider)
  └── infrastructure/
      └── scrapper/        # Letterboxd scraper
```

## Architecture

- **Hexagonal Architecture** - Separation of concerns
- **Ports & Adapters** - ReviewProvider abstraction for scraper
- **Stateless REST API** - Simple and fast
- **Gocolly Scraper** - Web scraping with rate limiting and user-agent spoofing

