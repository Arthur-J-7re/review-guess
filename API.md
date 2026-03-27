# Review Guess - Simple Backend API

A simple Go backend API for fetching Letterboxd reviews and managing a movie review guessing game. The API is designed to be consumed by any frontend (web, mobile, etc.).

## API Endpoints

### Health Check
```
GET /health
```
Returns basic API status.

**Response:**
```json
{
  "success": true,
  "data": "Review Guess API v1.0"
}
```

---

### Fetch Reviews
```
GET /api/reviews?username={username}
```

Fetches all reviews from a Letterboxd user.

**Parameters:**
- `username` (required): Letterboxd username

**Response:**
```json
{
  "success": true,
  "data": {
    "count": 18,
    "reviews": [
      {
        "author": "66sceptre",
        "title": "Alter Ego",
        "slug": "alter-ego-2026",
        "content": "Je me suis fait 15 fois...",
        "rating": 3,
        "liked": true,
        "spoilers": false
      }
    ]
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "No reviews found for this user"
}
```

---

### Start a Game
```
POST /api/game/start
```

Initializes a new game session by fetching reviews from specified users.

**Request Body:**
```json
{
  "usernames": ["66sceptre", "anderlybox"],
  "question_count": 10
}
```

**Parameters:**
- `usernames` (required): Array of Letterboxd usernames
- `question_count` (optional): Number of questions (default: 10)

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Game started",
    "total_questions": 10,
    "current_question": {
      "index": 1,
      "total": 10,
      "review": {
        "author": "66sceptre",
        "title": "Alter Ego",
        "slug": "alter-ego-2026",
        "content": "Je me suis fait 15 fois...",
        "rating": 3,
        "liked": true,
        "spoilers": false
      },
      "difficulty": 1.2
    }
  }
}
```

---

### Get Current Question
```
GET /api/game/question
```

Retrieves the current question without submitting an answer.

**Response:**
```json
{
  "success": true,
  "data": {
    "index": 1,
    "total": 10,
    "review": {
      "author": "66sceptre",
      "title": "Alter Ego",
      "slug": "alter-ego-2026",
      "content": "Je me suis fait 15 fois...",
      "rating": 3,
      "liked": true,
      "spoilers": false
    },
    "difficulty": 1.2
  }
}
```

---

### Submit an Answer
```
POST /api/game/answer
```

Submits the player's guess for the current question (author + film).

**Request Body:**
```json
{
  "guessed_author": "66sceptre",
  "guessed_film": "alter-ego-2026"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "correct": true,
    "partial_author": false,
    "partial_film": false,
    "correct_author": "66sceptre",
    "correct_film": "Alter Ego",
    "correct_slug": "alter-ego-2026",
    "points": 100,
    "current_score": 100,
    "is_game_over": false
  }
}
```

**Scoring:**
- Both correct (author + film): **100 points**
- Partial (either author OR film): **50 points**
- Both wrong: **0 points**

---

### Get Current Score
```
GET /api/game/score
```

Retrieves the current game score and progress.

**Response:**
```json
{
  "success": true,
  "data": {
    "current_score": 150,
    "answered": 2,
    "total_questions": 10,
    "is_game_over": false
  }
}
```

---

### Get Final Results
```
GET /api/game/results
```

Retrieves final results when game is complete. Only available after all questions answered.

**Response:**
```json
{
  "success": true,
  "data": {
    "score": 750,
    "total_points": 1000,
    "percentage": 75,
    "grade": "B",
    "answered": 10,
    "total": 10
  }
}
```

**Grade Scale:**
- A+: 90-100%
- A: 80-89%
- B: 70-79%
- C: 60-69%
- F: < 60%

---

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

**Common HTTP Status Codes:**
- `200 OK`: Request successful
- `400 Bad Request`: Invalid input or bad request
- `404 Not Found`: No reviews found
- `500 Internal Server Error`: Server error during execution
- `405 Method Not Allowed`: Wrong HTTP method

---

## Game Flow Example

1. **Start a game:**
   ```bash
   curl -X POST http://localhost:8080/api/game/start \
     -H "Content-Type: application/json" \
     -d '{
       "usernames": ["66sceptre"],
       "question_count": 5
     }'
   ```

2. **Get current question:**
   ```bash
   curl http://localhost:8080/api/game/question
   ```

3. **Submit answer:**
   ```bash
   curl -X POST http://localhost:8080/api/game/answer \
     -H "Content-Type: application/json" \
     -d '{
       "guessed_author": "66sceptre",
       "guessed_film": "alter-ego-2026"
     }'
   ```

4. **Repeat steps 2-3 for each question**

5. **Get final results:**
   ```bash
   curl http://localhost:8080/api/game/results
   ```

---

## Running the Server

```bash
go build -o review-guess ./cmd/review-guess
./review-guess
```

The server starts on `http://localhost:8080`

---

## Architecture

- **Domain Layer**: Core game models (Review, Film, Question, etc.)
- **Application Layer**: GameService with game logic
- **Infrastructure Layer**: Gocolly-based Letterboxd scraper
- **Adapter Layer**: HTTP handlers for REST API

---

## Notes

- Reviews are fetched from Letterboxd in real-time
- Rate limiting: 2 seconds between pages + 1 second random delay
- Questions are randomly shuffled from all fetched reviews
- Difficulty is calculated based on review length and rating presence
- Game state is stored in-memory per session
- Each game session is independent (creates new GameService)
