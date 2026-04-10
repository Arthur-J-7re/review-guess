# API Endpoints - Review Guess v2.0

## Base URL
```
http://localhost:8080
```

---

## Health Check

### GET /health
Returns API status.

**Response:**
```json
{
  "success": true,
  "data": "Review Guess API v2.0 - Player/Reviewer Architecture"
}
```

---

## Scraper Endpoints

### GET /api/reviews
Fetch reviews for one or multiple Letterboxd usernames.

**Query Parameters:**
- `username` (required, comma-separated) - Letterboxd username(s)

**Examples:**
```bash
# Single username
GET /api/reviews?username=alice

# Multiple usernames (comma-separated)
GET /api/reviews?username=alice,bob,charlie
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "review-1",
      "title": "Amazing film!",
      "movie_title": "The Shawshank Redemption",
      "username": "alice",
      "rating": 5.0,
      "content": "Best movie ever..."
    }
  ]
}
```

---

## Reviewer Management

### GET /api/reviewers
List all scraped reviewers in the system.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "reviewer-alice",
      "letterboxd_username": "alice",
      "total_reviews": 150,
      "total_movies_watched": 250,
      "last_scrapped_at": "2026-04-03T10:22:39Z",
      "created_at": "2026-04-02T14:30:00Z",
      "updated_at": "2026-04-03T10:22:39Z"
    }
  ]
}
```

---

### GET /api/reviewers/{username}
Get a specific reviewer by Letterboxd username.

**Path Parameters:**
- `username` (required) - Letterboxd username to retrieve

**Example:**
```bash
GET /api/reviewers/alice
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "reviewer-alice",
    "letterboxd_username": "alice",
    "total_reviews": 150,
    "total_movies_watched": 250,
    "last_scrapped_at": "2026-04-03T10:22:39Z",
    "created_at": "2026-04-02T14:30:00Z",
    "updated_at": "2026-04-03T10:22:39Z"
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "error": "Reviewer not found"
}
```

---

## Quiz Endpoints

### GET /api/quiz/next
Get the next quiz question for a player.

**Query Parameters:**
- `player_id` (required) - Unique player identifier

**Example:**
```bash
GET /api/quiz/next?player_id=player-alice-123
```

**Response:**
```json
{
  "success": true,
  "data": {
    "review_id": "review-456",
    "correct_movie_id": "tt0111161",
    "title": "One of the best films ever made",
    "content": "The story of two imprisoned men...",
    "options": [
      "tt0111161",
      "tt0068646",
      "tt0050083",
      "tt0047478"
    ]
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "error": "No quiz questions available"
}
```

---

### POST /api/quiz/answer
Record a player's quiz answer.

**Form Parameters:**
- `player_id` (required) - Player identifier
- `review_id` (required) - Review identifier (from quiz question)
- `answer_id` (optional) - Movie ID selected by player (empty string = skipped)

**Example:**
```bash
curl -X POST http://localhost:8080/api/quiz/answer \
  -d "player_id=player-alice-123" \
  -d "review_id=review-456" \
  -d "answer_id=tt0111161"
```

**Response (Correct):**
```json
{
  "success": true,
  "data": {
    "correct": true,
    "correct_movie_id": "tt0111161"
  }
}
```

**Response (Incorrect):**
```json
{
  "success": true,
  "data": {
    "correct": false,
    "correct_movie_id": "tt0111161"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "player_id and review_id are required"
}
```

---

### GET /api/quiz/stats
Get player's quiz statistics.

**Query Parameters:**
- `player_id` (required) - Player identifier

**Example:**
```bash
GET /api/quiz/stats?player_id=player-alice-123
```

**Response:**
```json
{
  "success": true,
  "data": {
    "player_id": "player-alice-123",
    "total": 10,
    "correct": 7,
    "accuracy": 70.0
  }
}
```

**Error Response (404):**
```json
{
  "success": false,
  "error": "Player not found or no answers recorded"
}
```

---

## HTTP Status Codes
- `200 OK` - Request succeeded
- `400 Bad Request` - Missing or invalid parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

