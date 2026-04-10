# Architecture Refactorée - Players vs Reviewers

## 🎯 Concept Clé

Le système separa maintenant **joueurs du quiz** et **reviewers Letterboxd**:

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  PLAYERS (Joueurs du quiz)                                 │
│  ├─ ID unique, nickname optionnel                          │
│  ├─ is_logged_in (bool)                                    │
│  ├─ Quiz history & scores                                  │
│  └─ Peuvent jouer sans compte                              │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  LETTERBOXD_REVIEWERS (Comptes Letterboxd)                │
│  ├─ letterboxd_username (le vrai compte)                  │
│  ├─ Reviews & films regardés                              │
│  ├─ Stats de scraping (pages, dates)                      │
│  └─ Indépendant des joueurs                               │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  PLAYER_REVIEWER_LINKS (Optionnel)                         │
│  ├─ Lie un joueur à un ou plusieurs reviewers              │
│  ├─ is_primary: quel reviewer utiliser défaut             │
│  └─ Permet un joueur d'avoir plusieurs reviewers           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Schéma Simplifié

```
PLAYERS
  ├─ id (PK)
  ├─ nickname (optionnel)
  ├─ is_logged_in (bool)
  ├─ last_reviewed_reviewer_id (FK reviewers)
  └─ created_at, last_played_at

LETTERBOXD_REVIEWERS
  ├─ id (PK)
  ├─ letterboxd_username (UNIQUE)
  ├─ total_reviews, total_movies_watched
  ├─ last_review_page_scrapped, last_movie_page_scrapped
  ├─ last_scrapped_at
  └─ created_at, updated_at

PLAYER_REVIEWER_LINKS
  ├─ player_id (FK players)
  ├─ reviewer_id (FK letterboxd_reviewers)
  ├─ is_primary (bool)
  └─ linked_at

REVIEWS
  ├─ id (PK)
  ├─ reviewer_id (FK letterboxd_reviewers) ← CHANGE
  ├─ movie_id (FK movies)
  ├─ content, rating, usable
  └─ ...

REVIEWER_MOVIES (ex: USER_MOVIES)
  ├─ reviewer_id (FK letterboxd_reviewers) ← CHANGE
  ├─ movie_id (FK movies)
  ├─ has_review, rating
  └─ ...

QUIZ_HISTORY
  ├─ id (PK)
  ├─ player_id (FK players) ← CHANGE
  ├─ review_id (FK reviews)
  ├─ player_answer_id (FK movies) ← CHANGE
  └─ ...
```

## 🔄 Cas d'Usage

### Cas 1: Visiteur Anonyme Jouant au Quiz
```
1. Player créé sans compte (is_logged_in=false)
2. Joue avec un reviewer aléatoire
3. Quiz history enregistré avec son player_id
4. Pas de lien permanent
```

### Cas 2: Joueur Connecté avec Son Compte Letterboxd
```
1. Player créé (is_logged_in=true)
2. Les reviews sont liées au reviewer_id du joueur
3. PlayerReviewerLink crée automatiquement (is_primary=true)
4. Quiz history tracé sur son player_id
5. Il voit uniquement SES reviews
```

### Cas 3: Joueur Explores Plusieurs Reviewers
```
1. Player a plusieurs PlayerReviewerLink
2. last_reviewed_reviewer_id = quel reviewer joue maintenant
3. Peut changer de reviewer dynamiquement
4. All quiz history encore lié à son player_id
```

### Cas 4: Pur Scraper - Pas de Joueur
```
1. LetterboxdReviewer crée/mis à jour par scraper
2. Reviews + ReviewerMovies remplies
3. Aucun player créé
4. Cet reviewer peut être utilisé par d'autres joueurs
```

## 🔌 Repositories Mis à Jour

### Nouveaux
- ✅ `PlayerRepository` - Joueurs
- ✅ `LetterboxdReviewerRepository` - Reviewers Letterboxd
- ✅ `PlayerReviewerLinkRepository` - Liens bidirectionnels

### Renommés/Modifiés
- ✅ `ReviewRepository`:
  - `GetByUserAndMovie()` → `GetByReviewerAndMovie()`
  - `GetUsableReviewsForUser()` → `GetUsableReviewsForReviewer()`

- ✅ `ReviewerMovieRepository` (était UserMovieRepository)
  - `Get(reviewerID, movieID)`
  - `GetMoviesWatchedByReviewer(reviewerID)`
  - `GetMoviesNotWatchedByReviewer(reviewerID, limit)`

- ✅ `QuizHistoryRepository`:
  - `Create(answer)` - use `PlayerID` + `PlayerAnswerID`
  - `GetPlayerHistory(playerID)`
  - `GetPlayerScores(playerID)`

## 💾 Exemple: Intégration dans main.go

```go
// Créer un joueur anonyme
player := &domain.Player{
	ID:         generateID(),
	IsLoggedIn: false,
	// Pas de reviewer spécifique
}
playerRepo.Create(player)

// Obtenir les reviews d'un reviewer
reviewer, _ := reviewerRepo.GetByUsername("alice")
reviews, _ := reviewRepo.GetUsableReviewsForReviewer(reviewer.ID)

// Jouer au quiz
answer := &domain.QuizAnswer{
	PlayerID:       player.ID,
	ReviewID:       review.ID,
	CorrectMovieID: movie.ID,
	PlayerAnswerID: &selectedMovieID,
	IsCorrect:      selectedMovieID == movie.ID,
}
quizRepo.Create(answer)

// Stats du joueur
total, correct, _ := quizRepo.GetPlayerScores(player.ID)
```

## 🚀 Avantages de cette Architecture

1. **Flexibilité**: Joueurs anonymes OU connectés
2. **Scalabilité**: Reviewers indépendants des joueurs
3. **Multiplexing**: Un joueur peut jouer avec plusieurs reviewers
4. **Indépendance**: Scraper fonctionne indépendamment du quiz
5. **Analytics**: Tracking séparé: joueur vs reviewer

## 📋 Fichiers Modifiés

✅ `migrations/001_init_schema.sql`
✅ `internal/domain/models.go`
✅ `internal/domain/ports.go`
✅ `internal/infrastructure/database/player_and_reviewer_repositories.go` (NEW)
✅ `internal/infrastructure/database/review_repository.go`
✅ `internal/infrastructure/database/other_repositories.go` (ReviewerMovieRepository)
✅ `internal/infrastructure/database/quiz_and_person_repositories.go`

## ⚠️ Breaking Changes

Si vous aviez du code utilisant l'ancienne `UserRepository`:
- `User` → `Player` + `LetterboxdReviewer`
- `user_id` → `player_id` ou `reviewer_id`
- `user_movies` → `reviewer_movies`
- `GetReviewsForUser()` → `GetReviewsForReviewer()`

**Toutes les migrations et implémentations sont à jour !**
