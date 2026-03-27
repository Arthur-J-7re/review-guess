# 🎬 Review Guess - Backend API

A simple Go backend API for fetching Letterboxd reviews and managing a movie review guessing game. Designed to be consumed by any frontend (web, mobile, CLI tools, etc.).

## Features

- 🌐 Real-time Letterboxd review scraping via Gocolly
- 🎮 Stateless game API ready for frontend integration
- 🏆 Scoring system (100pts for both correct, 50pts for partial, 0pts for wrong)
- 📊 Grade-based results display
- 🔄 Support for multiple user reviews per game
- ⚡ Lightweight and fast

## Prerequisites

- Go 1.24+
- Port 8080 available

## Installation

```bash
cd review-guess
go mod tidy
```

## Running the Server

### Development

```bash
go run ./cmd/review-guess
```

### Build Binary

```bash
go build -o review-guess ./cmd/review-guess
./review-guess
```

The API will start on `http://localhost:8080`

## API Overview

### Health Check
```bash
curl http://localhost:8080/health
```

### Fetch User Reviews
```bash
curl "http://localhost:8080/api/reviews?username=66sceptre"
```

### Start a Game
```bash
curl -X POST http://localhost:8080/api/game/start \
  -H "Content-Type: application/json" \
  -d '{
    "usernames": ["66sceptre", "anderlybox"],
    "question_count": 10
  }'
```

### Get Current Question
```bash
curl http://localhost:8080/api/game/question
```

### Submit an Answer
```bash
curl -X POST http://localhost:8080/api/game/answer \
  -H "Content-Type: application/json" \
  -d '{
    "guessed_author": "66sceptre",
    "guessed_film": "alter-ego-2026"
  }'
```

### Get Score
```bash
curl http://localhost:8080/api/game/score
```

### Get Final Results
```bash
curl http://localhost:8080/api/game/results
```

## Full API Documentation

See [API.md](API.md) for complete endpoint documentation, request/response examples, and error handling.

## Project Structure

```
review-guess/
├── cmd/review-guess/                 # API entry point
├── internal/
│   ├── domain/                       # Models & interfaces
│   ├── application/                  # GameService business logic
│   ├── infrastructure/
│   │   └── scrapper/                 # Letterboxd scraper
│   └── adapters/
│       └── httpapi/                  # REST API handlers
├── tests/                            # Test files
├── API.md                            # API documentation
├── go.mod
├── README.md
└── Taskfile.yml
```

## Architecture

- **Hexagonal Architecture** - Separation of concerns
- **Ports & Adapters** - ReviewProvider abstraction for scraper
- **Stateless REST API** - Each game session is independent
- **Gocolly Scraper** - Web scraping with rate limiting

## Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/api/reviews?username=...` | Fetch user reviews |
| POST | `/api/game/start` | Start new game |
| GET | `/api/game/question` | Get current question |
| POST | `/api/game/answer` | Submit answer |
| GET | `/api/game/score` | Get current score |
| GET | `/api/game/results` | Get final results |

## Scoring System

- **100 points** - Correct author AND film ✓✓
- **50 points** - Either correct author OR film ✓
- **0 points** - Both wrong ✗✗

Default max: 100 points per question

## Running Tests

```bash
go test ./...
```

## Development

### Build & Run
```bash
go build -o review-guess ./cmd/review-guess
./review-guess
```

### With Taskfile
```bash
task build
task run
task test
```

## Notes

- Reviews fetched from Letterboxd with real-time scraping
- Rate limiting: 2s per page + random 1s delay
- Quality filtering: minimum 30 characters per review
- Game state is in-memory (no persistence)
- Each game session is independent and isolated
- API is stateless - suitable for multi-user deployment

## Frontend Integration

This is designed as a backend-only API. You can build frontends in:
- React/Next.js (Web)
- Flutter/React Native (Mobile)
- Vue/Svelte (Web)
- Custom CLI tools
- Any technology that supports HTTP/JSON

## Future Enhancements

- [ ] Game history & statistics persistence
- [ ] Difficulty levels 
- [ ] Multi-player competitive mode
- [ ] User authentication
- [ ] Review filtering options
- [ ] Customizable scoring rules
- [ ] Database persistence

## License

MIT

## Author

Built with ❤️ using Go
