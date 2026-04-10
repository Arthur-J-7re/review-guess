# Database Infrastructure Setup

Ce répertoire contient toute l'infrastructure de base de données pour le système de quiz de films.

## Structure

```
internal/infrastructure/database/
├── client.go                          # Client DB singleton
├── movie_repository.go               # Film CRUD
├── user_repository.go                # Utilisateur CRUD
├── review_repository.go              # Review CRUD
├── other_repositories.go             # Similarité + UserMovie
├── quiz_and_person_repositories.go   # Quiz History + People
```

## Initialisation

### 1. Dépendances (ajouter au go.mod)

```bash
go get github.com/mattn/go-sqlite3
```

Pour PostgreSQL (optionnel) :
```bash
go get github.com/lib/pq
```

### 2. Créer le client DB au démarrage

**cmd/review-guess/main.go** :

```go
package main

import (
	"log"
	"net/http"

	"review-guess/internal/adapters/httpapi"
	"review-guess/internal/application"
	"review-guess/internal/infrastructure/database"
	"review-guess/internal/infrastructure/scrapper"
)

func main() {
	// === DATABASE ===
	// SQLite pour développement
	db, err := database.NewSQLiteClient("./review-guess.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// === REPOSITORIES ===
	sqlDB := db.GetDB()
	movieRepo := database.NewMovieRepository(sqlDB)
	userRepo := database.NewUserRepository(sqlDB)
	reviewRepo := database.NewReviewRepository(sqlDB)
	
	// === EXISTING CODE ===
	provider := scrapper.NewLetterboxdScraper()
	reviewService := application.NewReviewService(provider)

	// === ROUTER ===
	router := httpapi.NewRouter(reviewService)

	// Start server
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
```

## API d'utilisation

### Créer/Récupérer des films

```go
import "review-guess/internal/infrastructure/database"

movieRepo := database.NewMovieRepository(db)

// Créer
movie := &domain.Movie{
	ID: "tt0111161",
	Title: "The Shawshank Redemption",
	Year: 1994,
	LetterboxdSlug: "the-shawshank-redemption",
	TmdbID: 278,
}
err := movieRepo.Create(movie)

// Récupérer
movie, err := movieRepo.Get("tt0111161")

// Récupérer par TMDB
movie, err := movieRepo.GetByTmdbID(278)

// Récupérer plusieurs
movies, err := movieRepo.GetMany([]string{"id1", "id2", "id3"})
```

### Créer/Récupérer des utilisateurs

```go
userRepo := database.NewUserRepository(db)

// Créer
user := &domain.User{
	ID: "user123",
	LetterboxdUsername: "alice",
	TotalReviews: 50,
}
err := userRepo.Create(user)

// Récupérer par username
user, err := userRepo.GetByUsername("alice")

// Mettre à jour
user.TotalReviews = 51
err := userRepo.Update(user)
```

### Créer/Récupérer des reviews

```go
reviewRepo := database.NewReviewRepository(db)

// Créer
review := &domain.Review{
	ID: "review123",
	UserID: "user123",
	MovieID: "tt0111161",
	Title: "Amazing film!",
	Content: "Best movie ever...",
	Rating: 5.0,
	Usable: true,  // Peut être utilisée dans le quiz
}
err := reviewRepo.Create(review)

// Récupérer les reviews utilisables d'un user
reviews, err := reviewRepo.GetUsableReviewsForUser("user123")

// Récupérer une review aléatoire pour le quiz
randomReview, err := reviewRepo.GetRandomUsableReview()
```

### Créer par batch (optimisé)

```go
var reviews []*domain.Review
for _, letterboxdReview := range scrapedReviews {
	reviews = append(reviews, &domain.Review{
		ID: generateID(),
		UserID: "user123",
		MovieID: letterboxdReview.MovieID,
		Content: letterboxdReview.Content,
		Rating: letterboxdReview.Rating,
		Usable: true,
	})
}

err := reviewRepo.CreateBatch(reviews)  // Une seule transaction
```

### Similarité entre films

```go
simRepo := database.NewMovieSimilarityRepository(sqlDB)

// Créer relation
similarity := &domain.MovieSimilarity{
	MovieAID: "tt0111161",
	MovieBID: "tt0068646",
	SimilarityScore: 0.85,
	SharedDirectors: 0,
	SharedActors: 2,
	SharedGenres: 1,
	YearProximity: 1,
}
err := simRepo.Create(similarity)

// Récupérer les 5 films les plus similaires
similar, err := simRepo.GetTopSimilarMovies("tt0111161", 5)
for _, sim := range similar {
	fmt.Printf("Film similaire: %s (score: %.2f)\n", sim.MovieBID, sim.SimilarityScore)
}
```

### Historique utilisateur (films regardés)

```go
userMovieRepo := database.NewUserMovieRepository(sqlDB)

// Tracer un film regardé par un utilisateur
userMovie := &domain.UserMovie{
	UserID: "user123",
	MovieID: "tt0111161",
	HasReview: true,
	Rating: 5.0,
}
err := userMovieRepo.Create(userMovie)

// Récupérer tous les films regardés
watched, err := userMovieRepo.GetMoviesWatchedByUser("user123")

// Récupérer des films NON regardés par l'utilisateur (pour les leurres)
unwatched, err := userMovieRepo.GetMoviesNotWatchedByUser("user123", 100)
```

### Historique quiz

```go
quizRepo := database.NewQuizHistoryRepository(sqlDB)

// Enregistrer une réponse
answer := &domain.QuizAnswer{
	ID: "answer123",
	UserID: "user123",
	ReviewID: "review123",
	CorrectMovieID: "tt0111161",
	UserAnswerID: "tt0068646",
	IsCorrect: false,
	Options: []string{"tt0111161", "tt0068646", "tt0050083", "tt0047478"},
}
err := quizRepo.Create(answer)

// Récupérer l'historique d'un utilisateur
history, err := quizRepo.GetUserHistory("user123")

// Récupérer les scores
totalAnswered, correctAnswers, err := quizRepo.GetUserScores("user123")
accuracy := float64(correctAnswers) / float64(totalAnswered) * 100
fmt.Printf("Score: %d/%d (%.1f%%)\n", correctAnswers, totalAnswered, accuracy)
```

## Transactions

Pour opérations multi-tables :

```go
tx, err := db.BeginTx()
if err != nil {
	log.Fatal(err)
}
defer tx.Rollback()

// Plusieurs opérations...

if err := tx.Commit(); err != nil {
	log.Fatal(err)
}
```

## Performance

### Indexes créés automatiquement

- `reviews(user_id)` - Requêtes par utilisateur
- `reviews(usable)` - Requêtes pour le quiz
- `user_movies(user_id)` - Films regardés par user
- `movie_similarities(movie_a, similarity_score DESC)` - Top N similaires
- Et d'autres...

### Optimisations batch

Utiliser `CreateBatch()` au lieu de boucles :

```go
// ❌ Lent (N requêtes)
for _, review := range reviews {
	reviewRepo.Create(review)
}

// ✅ Rapide (1 transaction)
reviewRepo.CreateBatch(reviews)
```

## TODO - Repositories manquants

Certaines interfaces ne ont pas encore d'implémentations :
- `PersonRepository.Create()` avec movie_cast/movie_crew
- Intégration complète TMDBAPI
- Calcul des similarités (algorithme)

Ces services peuvent être ajoutés dans `application/` layer.

## Migration PostgreSQL

Pour passer à PostgreSQL en production :

```go
// Créer un client PostgreSQL
import "github.com/lib/pq"

db, err := sql.Open("postgres", "postgres://user:pass@localhost/review_guess")
```

Le schéma SQL fonctionne avec les deux !
