# 🎬 Review Guess - Quiz System

A Go backend API for a film quiz system that uses Letterboxd reviews to generate questions with coherent lures.

## 🎯 Features

- **Quiz System** - Players answer questions based on film reviews
- **Letterboxd Integration** - Scrape reviews from Letterboxd accounts
- **Player Tracking** - Track player scores (anonymous or registered)
- **Reviewer Management** - Manage scraped Letterboxd reviewer accounts
- **SQLite Database** - Persistent data storage with auto-migrations

---

## 📋 Prerequisites

- **Go 1.26+**
- **Docker & Docker Compose** (recommended) OR **GCC** (for local development with CGO)
- **Port 8080** available

---

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
# Build and run with Docker Compose
docker-compose up --build

# Or run individual commands
docker build -t review-guess .
docker run -p 8080:8080 review-guess
```

The API will be available at `http://localhost:8080`

### Option 2: Local Development (with GCC installed)

```bash
# Install dependencies
go mod download

# Build
task build
# or: CGO_ENABLED=1 go build -o review-guess ./cmd/review-guess

# Run
task run
# or: CGO_ENABLED=1 go run ./cmd/review-guess
```

### Option 3: Local Build (without Docker/GCC)

⚠️ **Note:** SQLite requires CGO to work. Without it, you'll need to:
1. Install GCC (`apt install build-essential` on Ubuntu/Debian)
2. Use Docker
3. Switch to a pure-Go SQLite driver (requires code changes)

---

## 📚 API Documentation

Full API documentation available in [API.md](API.md)

### Key Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Scraper - Get reviews from Letterboxd
curl "http://localhost:8080/api/reviews?username=alice,bob"

# Reviewers - Manage data sources
curl http://localhost:8080/api/reviewers
curl "http://localhost:8080/api/reviewers/alice"

# Quiz - Play the game
curl "http://localhost:8080/api/quiz/next?player_id=player-1"
curl -X POST http://localhost:8080/api/quiz/answer \
  -d "player_id=player-1" \
  -d "review_id=review-xyz" \
  -d "answer_id=tt0111161"
curl "http://localhost:8080/api/quiz/stats?player_id=player-1"
```

---

## 🏗️ Architecture

**Player vs Reviewer Separation:**
- **Players** - Quiz participants (anonymous or registered)
- **Letterboxd Reviewers** - Scraped accounts (data sources)
- **Quiz History** - Track player answers and scores

**Hexagonal Architecture:**
```
Domain Layer (models, interfaces)
    ↓
Application Layer (business logic)
    ↓
Infrastructure Layer (database, scraper, HTTP)
    ↓
Adapters (HTTP API)
```

---

## 📁 Project Structure

```
cmd/review-guess/
  └── main.go                    # Application entry point

internal/
  ├── adapters/
  │   └── httpapi/             # HTTP routes and handlers
  ├── application/
  │   └── review_service.go    # Business logic
  ├── domain/
  │   ├── models.go            # Domain entities (Player, Review, Quiz, etc.)
  │   ├── ports.go             # Repository interfaces
  │   └── errors.go            # Error definitions
  └── infrastructure/
      ├── database/            # SQLite repositories
      └── scrapper/            # Letterboxd scraper

migrations/
  └── 001_init_schema.sql      # Database schema

API.md                          # Complete API documentation
DATABASE_REFACTORED.md          # Database schema reference
ARCHITECTURE_REFACTORED.md      # Architecture patterns
```

---

## 🗄️ Database

**SQLite** with automatic migrations on startup.

Default location: `./review-guess.db`

### Tables
- `players` - Quiz participants
- `letterboxd_reviewers` - Scraped reviewer accounts
- `reviews` - Film reviews from reviewers
- `movies` - Film metadata
- `quiz_history` - Player quiz answers and scores

See [DATABASE_REFACTORED.md](DATABASE_REFACTORED.md) for complete schema.

---

## 🔧 Development

### Available Tasks (using Task CLI)

```bash
task build      # Build binary with CGO
task run        # Run server with CGO
task clean      # Remove build artifacts
task deps       # Download/tidy dependencies
task fmt        # Format Go code
```

### Testing the API

```bash
# Terminal 1: Start server
task run

# Terminal 2: Test endpoints
curl http://localhost:8080/health
curl "http://localhost:8080/api/quiz/next?player_id=test-player"
curl "http://localhost:8080/api/quiz/stats?player_id=test-player"
```

---

## 🐳 Docker Commands

### Build
```bash
docker build -t review-guess .
```

### Run
```bash
docker run -p 8080:8080 review-guess
```

### With persistent database
```bash
docker run -p 8080:8080 -v $(pwd)/data:/app/data review-guess
```

### Using Docker Compose
```bash
docker-compose up              # Build and run
docker-compose up --build      # Rebuild and run
docker-compose down            # Stop and remove containers
docker-compose logs -f         # View logs
```

---

## �️ Database Management

When running with Docker Compose, SQLite Web Manager is automatically available for browsing and editing the database.

### Via Docker Compose
```bash
# Start the app with Docker Compose
docker-compose up
```

### SQLite Access
```bash
# If sqlite3 is installed locally
sqlite3 ./data/review-guess.db

# Useful commands:
.tables                          # List all tables
SELECT * FROM players;           # View all players
SELECT * FROM letterboxd_reviewers;  # View all reviewers
SELECT * FROM reviews;           # View all reviews
SELECT * FROM quiz_history;      # View quiz attempts
```

---

## �📖 Documentation

- [API.md](API.md) - Complete API reference
- [DATABASE_REFACTORED.md](DATABASE_REFACTORED.md) - Database schema
- [ARCHITECTURE_REFACTORED.md](ARCHITECTURE_REFACTORED.md) - System design patterns
- [REFACTORING_SUMMARY.md](REFACTORING_SUMMARY.md) - Player/Reviewer separation details

---

## 🆘 Troubleshooting

### Error: "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo"

**Solution:** Use `CGO_ENABLED=1` when building/running:
```bash
CGO_ENABLED=1 go build -o review-guess ./cmd/review-guess
CGO_ENABLED=1 go run ./cmd/review-guess
```

Or use Docker:
```bash
docker-compose up
```

### Error: gcc not found

**Solution:** Install GCC or use Docker:

**Ubuntu/Debian:**
```bash
apt install build-essential
```

**macOS:**
```bash
xcode-select --install
```

**Windows:**
Install [MinGW](https://www.mingw-w64.org/) or use [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install)

Or simply use Docker:
```bash
docker-compose up
```

---

## 🎓 Learning Resources

- [IMPLEMENTATION_CHECKLIST_PHASE4.md](IMPLEMENTATION_CHECKLIST_PHASE4.md) - Integration guide with code examples
- [ARCHITECTURE_REFACTORED.md](ARCHITECTURE_REFACTORED.md) - Design patterns and use cases

---

## 📝 License

MIT

