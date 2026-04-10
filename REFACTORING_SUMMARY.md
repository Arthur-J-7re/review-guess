# ⚡ Refactoring: Players vs Reviewers - Summary

## 🎯 Changement Principal

**Avant**: Une table `users` qui confondait joueurs du quiz et reviewers Letterboxd
**Après**: Deux entités distinctes:
- **`players`** - Joueurs du quiz (anonymes ou connectés)
- **`letterboxd_reviewers`** - Comptes Letterboxd scraped
- **`player_reviewer_links`** - Liens optionnels entre les deux

## 📝 Changements en Détail

### Models (`internal/domain/models.go`)
```diff
- User → REMOVED
+ Player (ID, nickname, is_logged_in, created_at, last_played_at)
+ LetterboxdReviewer (ID, letterboxd_username, stats)
+ PlayerReviewerLink (player_id, reviewer_id, is_primary)

- Review.user_id → Review.reviewer_id
- UserMovie → ReviewerMovie (reviewer_id)
- QuizAnswer.user_id → QuizAnswer.player_id
- QuizAnswer.user_answer_id → QuizAnswer.player_answer_id (nullable)
```

### Database Schema (`migrations/001_init_schema.sql`)
```diff
- users → REMOVED
+ players (id, nickname, is_logged_in, last_reviewed_reviewer_id, created_at, last_played_at)
+ letterboxd_reviewers (id, letterboxd_username, total_reviews, ...)
+ player_reviewer_links (player_id, reviewer_id, is_primary)

  reviews:
    - user_id → reviewer_id (FK letterboxd_reviewers)
    
  user_movies → reviewer_movies:
    - user_id → reviewer_id
```

### Interfaces (`internal/domain/ports.go`)
```diff
- UserRepository → REMOVED
+ PlayerRepository (Get, Create, Update, List)
+ LetterboxdReviewerRepository (Get, GetByUsername, Create, Update, List)
+ PlayerReviewerLinkRepository (Create, Get, GetPlayerReviewers, Delete)

  ReviewRepository:
    - GetByUserAndMovie() → GetByReviewerAndMovie()
    - GetUsableReviewsForUser() → GetUsableReviewsForReviewer()
    
  UserMovieRepository → ReviewerMovieRepository:
    - Get(userID, movieID) → Get(reviewerID, movieID)
    - GetMoviesWatchedByUser() → GetMoviesWatchedByReviewer()
    - GetMoviesNotWatchedByUser() → GetMoviesNotWatchedByReviewer()
    
  QuizHistoryRepository:
    - GetUserHistory() → GetPlayerHistory()
    - GetUserScores() → GetPlayerScores()
```

### Implementations (`internal/infrastructure/database/`)
```
NEW: player_and_reviewer_repositories.go
  ✅ PlayerRepositoryImpl
  ✅ LetterboxdReviewerRepositoryImpl
  ✅ PlayerReviewerLinkRepositoryImpl

UPDATED: review_repository.go
  ✅ Changed: user_id → reviewer_id in all queries

UPDATED: other_repositories.go
  ✅ Renamed: UserMovieRepository → ReviewerMovieRepository
  ✅ Changed: user_id/userID → reviewer_id/reviewerID in all methods

UPDATED: quiz_and_person_repositories.go
  ✅ Changed: user_id → player_id
  ✅ Changed: user_answer_id → player_answer_id
  ✅ Renamed: GetUserHistory → GetPlayerHistory
  ✅ Renamed: GetUserScores → GetPlayerScores
```

## 🎮 Cas d'Usage

### 1. Visiteur Anonyme
```go
// Pas de compte - joue une fois
player := &domain.Player{
	ID:         generateID(),
	IsLoggedIn: false,
}
playerRepo.Create(player)

// Utilise un reviewer aléatoire
reviews, _ := reviewRepo.GetRandomUsableReview()
```

### 2. Joueur Connecté
```go
// Créer joueur + lier à son reviewer Letterboxd
player := &domain.Player{
	ID:         generateID(),
	IsLoggedIn: true,
	Nickname:   "Alice",
}
playerRepo.Create(player)

// Trouver son reviewer
reviewer, _ := reviewerRepo.GetByUsername("alice")

// Lier
link := &domain.PlayerReviewerLink{
	PlayerID:   player.ID,
	ReviewerID: reviewer.ID,
	IsPrimary:  true,
}
linkRepo.Create(link)

// Jouer uniquement avec SES reviews
reviews, _ := reviewRepo.GetUsableReviewsForReviewer(reviewer.ID)
```

### 3. Scraper (Indépendant)
```go
// Crée/met à jour un reviewer - aucun joueur impliqué
reviewer := &domain.LetterboxdReviewer{
	ID:                    "bob-id",
	LetterboxdUsername:    "bob",
	TotalReviews:          50,
	LastReviewPageScrapped: 5,
}
reviewerRepo.Create(reviewer)

// Importe les reviews
reviews := []*domain.Review{...}
reviewRepo.CreateBatch(reviews)
```

## ✅ What's Included

✅ Updated migrations with new tables
✅ Updated domain models (Player, LetterboxdReviewer)
✅ Updated port interfaces (separate repos)
✅ 3 new repository implementations
✅ 4 updated repository implementations
✅ New ARCHITECTURE_REFACTORED.md

## 🚀 Next: Integration

1. **Build & Test**
   ```bash
   go build ./cmd/review-guess
   ```

2. **Update main.go** to use new repos:
   ```go
   playerRepo := database.NewPlayerRepository(db)
   reviewerRepo := database.NewLetterboxdReviewerRepository(db)
   linkRepo := database.NewPlayerReviewerLinkRepository(db)
   ```

3. **Update quiz endpoints** to use:
   - `playerRepo` for player tracking
   - `reviewerRepo` for quiz sources
   - `linkRepo` for player preferences

## 💾 Migration Path

If upgrading existing DB:
```sql
-- Backup old users table
ALTER TABLE users RENAME TO users_backup;

-- New schema will create players & letterboxd_reviewers
-- Map old users → letterboxd_reviewers
INSERT INTO letterboxd_reviewers 
SELECT id, letterboxd_username, total_reviews, ... FROM users_backup;

-- Delete old data
DROP TABLE users_backup;
```

---

Architecture now elegantly separates concerns:
- **Players** = anyone playing the quiz
- **Reviewers** = Letterboxd data sources
- **Links** = optional connection between them

Perfect for your use case! 🎯
