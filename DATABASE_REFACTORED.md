# Database Schema Documentation (REFACTORED)

## Overview - Players vs Reviewers Architecture

The database now clearly separates **quiz players** from **data sources** (reviewers):

- **Players**: Quiz game participants (can be anonymous or logged-in)
- **Letterboxd Reviewers**: Data sources scraped from Letterboxd (independent accounts)
- **Player-Reviewer Links**: Optional connections when a logged-in player wants to use a specific reviewer
- **Movies**: Film data enriched from TMDB
- **Reviews**: Reviews from Letterboxd, tied to reviewers (not players)
- **Quiz History**: Player quiz answers and scores

This enables 4 use cases:
1. Anonymous player + random reviewer
2. Logged player + their own Letterboxd account replay
3. Logged player + multiple reviewers to choose from
4. Pure scraper (reviewer with no associated players)

## Tables

### `players`
Quiz game participants.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Unique player identifier |
| `nickname` | TEXT NULL | Display name (optional for anonymous) |
| `is_logged_in` | BOOLEAN | True if account holder, false if anonymous |
| `last_reviewed_reviewer_id` | TEXT NULL | Last reviewer they played with (FK) |
| `created_at` | TIMESTAMP | First played time |
| `last_played_at` | TIMESTAMP | Most recent quiz activity |
| `updated_at` | TIMESTAMP | Last modification |

**Use:**
- Anonymous visitor: `is_logged_in=FALSE, nickname=NULL`
- Logged player: `is_logged_in=TRUE, nickname="alice"`

### `letterboxd_reviewers`
Letterboxd accounts scraped for reviews.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Unique reviewer identifier |
| `letterboxd_username` | TEXT UNIQUE | Letterboxd handle (e.g., "alice") |
| `total_reviews` | INT | Reviews on Letterboxd |
| `total_movies_watched` | INT | Films watched |
| `last_review_page_scrapped` | INT | Pagination tracking for reviews |
| `last_movie_page_scrapped` | INT | Pagination tracking for watched films |
| `last_scrapped_at` | TIMESTAMP | Last data refresh |
| `created_at` | TIMESTAMP | When first added |
| `updated_at` | TIMESTAMP | Last modification |

**Use:**
- Represents independent Letterboxd data source
- Can exist without any player linked

### `player_reviewer_links`
Connections between players and reviewers (optional).

| Column | Type | Description |
|--------|------|-------------|
| `player_id` | TEXT | FK to `players` |
| `reviewer_id` | TEXT | FK to `letterboxd_reviewers` |
| `is_primary` | BOOLEAN | True if primary reviewer choice |
| `linked_at` | TIMESTAMP | When connection created |

**Constraints:**
- `PRIMARY KEY (player_id, reviewer_id)`

**Use:**
- Logged player "alice" links to their Letterboxd reviewer account "alice_lbx"
- Logged player can have multiple reviewers, one marked primary
- Anonymous players never appear here

### `movies`
Film data enriched from TMDB.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Internal ID |
| `title` | TEXT NOT NULL | Movie title |
| `year` | INT | Release year |
| `description` | TEXT | Plot synopsis |
| `poster_url` | TEXT | URL to poster |
| `tmdb_id` | INT UNIQUE | TMDB database ID |
| `letterboxd_slug` | TEXT UNIQUE | Letterboxd URL slug |
| `created_at` | TIMESTAMP | Row creation |
| `updated_at` | TIMESTAMP | Last update |

### `people`
Actors and directors.

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Internal ID |
| `tmdb_id` | INT UNIQUE | TMDB person ID |
| `name` | TEXT NOT NULL | Full name |
| `profile_picture` | TEXT | Profile image URL |
| `created_at` | TIMESTAMP | Row creation |

### `movie_cast`
Actors in each film.

| Column | Type | Description |
|--------|------|-------------|
| `movie_id` | TEXT | FK to `movies` |
| `person_id` | TEXT | FK to `people` |
| `character_name` | TEXT | Character name |
| `role_order` | INT | Importance (1=lead) |

### `movie_crew`
Directors and other crew.

| Column | Type | Description |
|--------|------|-------------|
| `movie_id` | TEXT | FK to `movies` |
| `person_id` | TEXT | FK to `people` |
| `job` | TEXT | Job title |

### `genres`
TMDB genres.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INT PRIMARY KEY | TMDB genre ID |
| `name` | TEXT UNIQUE | Genre name |

### `movie_genres`
Many-to-many movies ↔ genres.

| Column | Type | Description |
|--------|------|-------------|
| `movie_id` | TEXT | FK to `movies` |
| `genre_id` | INT | FK to `genres` |

### `reviews`
Letterboxd reviews (tied to **reviewers**, not players).

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Review ID |
| `reviewer_id` | TEXT NOT NULL | FK to `letterboxd_reviewers` |
| `movie_id` | TEXT NOT NULL | FK to `movies` |
| `title` | TEXT | Review title |
| `content` | TEXT | Review text |
| `rating` | FLOAT | Rating (0.5-5.0) |
| `liked` | BOOLEAN | Liked/favorited |
| `spoilers` | BOOLEAN | Contains spoilers |
| `usable` | BOOLEAN DEFAULT TRUE | Can be used in quiz |
| `created_at` | TIMESTAMP | Publication date |
| `updated_at` | TIMESTAMP | Last scrape update |

**Constraints:**
- `UNIQUE(reviewer_id, movie_id)` - One review per reviewer per film

### `reviewer_movies`
Films watched by each reviewer (from Letterboxd profile).

| Column | Type | Description |
|--------|------|-------------|
| `reviewer_id` | TEXT NOT NULL | FK to `letterboxd_reviewers` |
| `movie_id` | TEXT NOT NULL | FK to `movies` |
| `watched_at` | TIMESTAMP | When watched |

**Constraints:**
- `PRIMARY KEY (reviewer_id, movie_id)`

### `movie_similarities`
Pre-calculated similarity scores between films.

| Column | Type | Description |
|--------|------|-------------|
| `movie_a` | TEXT NOT NULL | First film ID |
| `movie_b` | TEXT NOT NULL | Second film ID |
| `similarity_score` | FLOAT | Score 0.0-1.0 |
| `shared_directors` | INT | Directors in common |
| `shared_actors` | INT | Actors in common |
| `shared_genres` | INT | Genres in common |
| `year_proximity` | INT | Year difference |
| `calculated_at` | TIMESTAMP | Calculation time |

**Constraints:**
- `PRIMARY KEY (movie_a, movie_b)`

### `quiz_history`
Quiz answers (tied to **players**, not reviewers).

| Column | Type | Description |
|--------|------|-------------|
| `id` | TEXT PRIMARY KEY | Record ID |
| `player_id` | TEXT NOT NULL | FK to `players` |
| `review_id` | TEXT NOT NULL | FK to `reviews` |
| `correct_movie_id` | TEXT NOT NULL | Right answer |
| `player_answer_id` | TEXT NULL | What player selected |
| `is_correct` | BOOLEAN | Correct/incorrect |
| `options` | TEXT (JSON) | Array of movie IDs |
| `answered_at` | TIMESTAMP | Answer time |

**Use:**
- Tracks per-player quiz performance
- Links to review (which ties to reviewer)
- `player_answer_id` is NULL if player skipped

## Initialization

### SQLite (Development)
```bash
# Auto-runs on first database.NewSQLiteClient("review-guess.db") call
# Or manually:
sqlite3 review-guess.db < migrations/001_init_schema.sql
```

### PostgreSQL (Production)
```bash
psql -U postgres -d review_guess < migrations/001_init_schema.sql
```

## Key Indexes

- `idx_players_created_at` - Find players by signup time
- `idx_players_last_played_at` - Recent players first
- `idx_letterboxd_reviewers_created_at` - Reviewers by add time
- `idx_letterboxd_reviewers_updated_at` - Refresh order
- `idx_reviews_reviewer_id` - Reviews by source
- `idx_reviews_usable` - Find usable quiz questions
- `idx_reviewer_movies_reviewer_id` - Films watched by reviewer
- `idx_movie_similarities_movie_a` - Similar films
- `idx_quiz_history_player_id` - Player's quiz history

## Data Flow

```
1. SCRAPER: Fetches from Letterboxd
   → Creates/updates letterboxd_reviewers
   → Creates reviews (tied to reviewer)
   → Updates reviewer_movies 
   
2. TMDB ENRICHER: Enriches movie data
   → Adds actors, directors, genres to movies
   
3. SIMILARITY CALCULATOR: Pre-calculates
   → movie_similarities table (can be async)
   
4. QUIZ ENGINE: Generates questions
   → Player plays quiz
   → Gets review from chosen/random reviewer
   → Finds similar movies as lures (using similarity_score)
   → Records answer in quiz_history
```

## Query Examples

### Find reviews for a specific reviewer
```sql
SELECT * FROM reviews 
WHERE reviewer_id = 'alice-id' AND usable = TRUE;
```

### Get player's quiz history
```sql
SELECT * FROM quiz_history 
WHERE player_id = 'player-123'
ORDER BY answered_at DESC;
```

### Find similar films to use as lures
```sql
SELECT movie_b FROM movie_similarities 
WHERE movie_a = 'tt0111161' 
ORDER BY similarity_score DESC LIMIT 4;
```

### Track which reviewers a player uses
```sql
SELECT r.* FROM letterboxd_reviewers r
JOIN player_reviewer_links l ON l.reviewer_id = r.id
WHERE l.player_id = 'player-123';
```

## Breaking Changes from Previous Schema

| Old | New |
|-----|-----|
| `users` table | Split into `players` + `letterboxd_reviewers` |
| `reviews.user_id` | → `reviews.reviewer_id` |
| `user_movies` | → `reviewer_movies` |
| `user_id` in quiz_history | → `player_id` |
| `user_answer_id` in quiz_history | → `player_answer_id` |

All code must be updated accordingly. See REFACTORING_SUMMARY.md and ARCHITECTURE_REFACTORED.md for migration details.
