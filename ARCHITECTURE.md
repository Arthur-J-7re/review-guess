# Review Guess API - Architecture

## Overview

A simple REST API for fetching Letterboxd reviews from one or multiple users.

## Architecture Pattern

**Hexagonal Architecture** with clear separation of concerns:

```
Presentation (HTTP)
    ↓
Application (Business Logic)
    ↓
Domain (Models)
    ↓
Infrastructure (Scraper)
```

## Components

### Domain Layer (`internal/domain/`)

**Models:**
- `Review`: Represents a single review
  - author, title, slug, content, rating, liked, spoilers
- `Reviews`: Collection of reviews
  - count, list of reviews

**Ports (Interfaces):**
- `ReviewProvider`: Interface for scraping reviews
  - `FetchUserReviews(username string) ([]*Review, error)`

**Errors:**
- `ErrInvalidUsername`: Invalid letterboxd username
- `ScrapperError`: Wrapped scraping errors

### Application Layer (`internal/application/`)

**ReviewService:**
- Orchestrates review fetching from multiple users
- Merges reviews from multiple sources
- Validates inputs

```go
type ReviewService struct {
    provider ReviewProvider
}

func (s *ReviewService) GetReviews(usernames ...string) (*Reviews, error)
```

### Infrastructure Layer (`internal/infrastructure/`)

**Scraper (`scrapper/`):**
- Implements `ReviewProvider` interface
- Uses Gocolly for web scraping
- Rate limiting (3s delay + 2s random)
- Browser-like headers for authentication
- Parses Letterboxd HTML structure

### Adapter Layer (`internal/adapters/`)

**HTTP API (`httpapi/`):**
- Router: Registers HTTP routes
- Handlers: HTTP handlers for endpoints
- Response formatting

**Routes:**
- `GET /health` - Health check
- `GET /api/reviews?username=...` - Fetch reviews

## Data Flow

```
1. HTTP Request
     ↓
2. Router (Route matching)
     ↓
3. Handler (Extract parameters)
     ↓
4. ReviewService (Business logic)
     ↓
5. Scraper (ReviewProvider implementation)
     ↓
6. Letterboxd (Web scraping)
     ↓
7. Response (JSON)
```

### Example Flow

```
GET /api/reviews?username=alice&username=bob
    ↓
Router → handleGetReviews
    ↓
Extract: usernames = ["alice", "bob"]
    ↓
ReviewService.GetReviews("alice", "bob")
    ↓
For each user:
  → Scraper.FetchUserReviews("alice")
  → Parse HTML, extract reviews
  → Return []*Review
    ↓
Merge all reviews
    ↓
Return Reviews{Count: X, Reviews: []}
    ↓
JSON Response
```

## Key Design Decisions

1. **Simplicity First**: Only what's needed for review fetching
2. **Stateless**: No persistence, no game logic
3. **Interface-Driven**: Scraper behind ReviewProvider interface for testability
4. **Rate-Limited**: Respectful web scraping with delays
5. **Error Handling**: Clear error types and wrapping
6. **Hexagonal**: Easy to test and extend

## Project Structure

```
review-guess/
├── cmd/
│   └── review-guess/
│       └── main.go              # Entry point
├── internal/
│   ├── domain/
│   │   ├── models.go            # Review, Reviews structs
│   │   ├── errors.go            # Error definitions
│   │   └── ports.go             # ReviewProvider interface
│   ├── application/
│   │   └── review_service.go    # ReviewService
│   ├── infrastructure/
│   │   └── scrapper/
│   │       └── scrapper.go      # Letterboxd scraper
│   └── adapters/
│       └── httpapi/
│           ├── router.go        # HTTP router
│           └── handlers.go      # HTTP handlers
├── API.md                       # API documentation
├── README.md
└── go.mod
```

## Dependencies

- `github.com/gocolly/colly/v2`: Web scraping library
- `github.com/charmbracelet/log`: Logging

## Future Extensibility

The architecture allows easy addition of:
- Caching layer
- Database persistence
- Different scraper implementations
- GraphQL endpoint
- WebSocket for real-time updates
- Additional review sources (IMDb, etc.)
    // unique
    Author    string    // username
    Film      *Film     // sur quel film
    Content   string    // le texte de la review
    Rating    int       // note 0-5 (0 = "watched" sans note)
    Liked     bool      // ❤️ ou pas
    Spoilers  bool      // marquée comme spoilers?
}

// Question = une question pendant le jeu
type Question struct {
    ReviewIndex  int    // idx dans la liste des reviews chargées
    Review       *Review
    Difficulty   float32 // calculé au "runtime"
}

// GameState = état du jeu en cours (en mémoire)
type GameState struct {
    Users       []*User      // tous les utilisateurs loadés
    AllReviews  []*Review    // fusion de toutes les reviews
    Questions   []*Question  // les questions de ce jeu
    Answers     []*Answer    // réponses du joueur
    Score       int
    CurrentIdx  int          // index question actuelle
}

// Answer = réponse du joueur à une question
type Answer struct {
    QuestionIdx  int
    GuessedUser  string // username deviné
    GuessedFilm  string // film slug deviné
    IsCorrectUser bool
    IsCorrectFilm bool
    Points        int
}
```

---

## 2️⃣ **Ports (Interfaces Métier)**

Fichier: `review-guess/internal/domain/ports.go`

**Système simple: un seul port, le scrapper**

```go
// ReviewProvider = Interface pour récupérer les reviews
type ReviewProvider interface {
    // Récupère toutes les reviews d'un utilisateur (toutes les pages)
    FetchUserReviews(username string) ([]*Review, error)
}
```

C'est tout! Pas besoin de Repository (tout en mémoire), pas besoin de GameService (logique dans un service simple).

---

## 3️⃣ **Services Applicatifs**

Fichier: `review-guess/internal/application/`

### **Juste UN service: GameService**

```go
// game_service.go

type GameService struct {
    scrapper ReviewProvider  // pour fetcher les reviews
    state    *GameState      // état du jeu (en mémoire)
}

// LoadGame: Charge les reviews des utilisateurs et prépare le jeu
func (s *GameService) LoadGame(usernames []string, questionCount int) error {
    // 1. Pour chaque username: FetchUserReviews()
    // 2. Fusionne toutes les reviews dans AllReviews
    // 3. Génère `questionCount` questions aléatoires filtrées
    // 4. Initialise GameState
}

// GetCurrentQuestion: Retourne la question courante (review masquée)
func (s *GameService) GetCurrentQuestion() (*Question, error) {
    // Retourne la question à l'idx courant
    // La review est masquée (Content laissé, mais pas auteur/film)
}

// SubmitAnswer: Valide la réponse et attribue points
func (s *GameService) SubmitAnswer(guessedUser string, guessedFilm string) (*Answer, error) {
    // 1. Compare avec la vraie réponse
    // 2. Attribue points (100 si les 2, 50 chacun si 1)
    // 3. Enregistre dans Answers
    // 4. Avance à la question suivante
}

// GetScore: Retourne le score actuel
func (s *GameService) GetScore() int {
    return s.state.Score
}

// IsGameOver: Vérifie si c'est fini
func (s *GameService) IsGameOver() bool {
    return s.state.CurrentIdx >= len(s.state.Questions)
}

// GetResults: Retourne récap final
func (s *GameService) GetResults() *GameResults {
    // Stats finales
}
```

**Utilitaires:**
```go
// difficulty.go - helper pour calculer la difficulté d'une review
func CalculateDifficulty(review *Review) float32 {
    // Short review = difficile
    // Long review = facile (plus d'indices)
    // No rating = x1.5 difficulté
}

// filtering.go - helper pour filtrer les bonnes questions
func FilterReviewsForGame(reviews []*Review) []*Review {
    // Enlève:
    //   - Reviews trop courtes (< 30 chars)
    //   - "Watched" sans contenu
    // Keep: Reviews avec contenu significatif
}
```

---

## 4️⃣ **Infrastructure - Scraping**

Fichier: `review-guess/internal/infrastructure/scrapper/`

**Ultralegersimple - juste le scrapper Letterboxd**

```
├── scrapper.go              # Factory + Letterboxd scraper
├── reviews.go               # Fetch reviews paginated
└── shared.go                # Helpers communs avec twin-pick
```

### **Scrapper Letterboxd pour reviews:**

```go
// scrapper.go
type Scrapper struct {
    baseURL string // https://letterboxd.com
}

func NewScrapper() *Scrapper {
    return &Scrapper{baseURL: "https://letterboxd.com"}
}

func (s *Scrapper) FetchUserReviews(username string) ([]*Review, error) {
    // Utilise Gocolly pour paginer les reviews
    // https://letterboxd.com/{username}/reviews/
    // https://letterboxd.com/{username}/reviews/page/2/
    // ...etc jusqu'à pas de résultats
    
    // Pour chaque page:
    //   - Parse film slug, title, year
    //   - Parse review content
    //   - Parse rating (★), liked (❤️), spoilers marker
    //   - Combine en []*Review
}
```

### **Rate limiting & etiquette:**
```
- Max 3-5 concurrent pages
- 1-2s delay entre requêtes
- User-Agent: Mozilla/5.0...
- Pas de cache (pas réutilisé)
```

---

## 5️⃣ **Adapters (Interfaces Utilisateur)**

### **Seul Adapter: CLI Interactive**
Fichier: `review-guess/cmd/review-guess/`

```go
// main.go

func main() {
    // 1. Affiche menu de démarrage
    fmt.Println("=== Review Guesser ===")
    
    // 2. Demande les pseudos (Cobra prompter ou simple Scanln)
    usernames := askForUsernames()
    questionCount := askQuestionCount()
    
    // 3. Fetch reviews (spinner de progression)
    game := &GameService{scrapper: NewScrapper()}
    if err := game.LoadGame(usernames, questionCount); err != nil {
        fmt.Println("Erreur:", err)
        return
    }
    
    // 4. Boucle du jeu
    for !game.IsGameOver() {
        question := game.GetCurrentQuestion()
        displayQuestion(question)  // Affiche la review masquée
        
        author := promptInput("Qui a écrit cette review? ")
        film := promptInput("Quel film? ")
        
        answer := game.SubmitAnswer(author, film)
        displayResult(answer)       // ✓/✗ + explique la réponse
        
        fmt.Printf("Score: %d\n", game.GetScore())
        promptContinue("Appuie pour la question suivante...")
    }
    
    // 5. Écran final
    displayFinalResults(game.GetResults())
}
```

**No HTTP API needed (unless later for UI web)**

---

## 6️⃣ **Structure des dossiers finale**

```
review-guess/
├── cmd/
│   └── review-guess/           # Entry point unique - CLI du jeu
│       └── main.go
│
├── internal/
│   ├── domain/
│   │   ├── models.go           # Review, Film, User, GameState
│   │   ├── ports.go            # ReviewProvider interface
│   │   └── errors.go
│   │
│   ├── application/
│   │   └── game_service.go     # SEUL service (LoadGame, SubmitAnswer, GetScore)
│   │
│   ├── infrastructure/
│   │   └── scrapper/
│   │       ├── scrapper.go     # Factory
│   │       ├── reviews.go      # FetchUserReviews (Gocolly)
│   │       └── shared.go       # Helpers, filtering
│   │
│   └── adapters/
│       └── cli/
│           ├── prompter.go     # Interactive prompts
│           ├── formatter.go    # Pretty output (colors, tables)
│           └── spinner.go      # Loading spinners
│
├── tests/
│   ├── fixtures/               # Test data (reviews examples)
│   └── mocks/                  # Mock scrapper
│
├── go.mod
├── go.sum
├── Taskfile.yml                # build, run, test
├── ARCHITECTURE.md             # Ce fichier
└── README.md                   # Guide rapide
```

**C'est tout!** Structure ultra-légère, zéro persistence, zéro sessions.

---

## 7️⃣ **Dépendances Go proposées**

```go
module review-guess

go 1.24

require (
    github.com/gocolly/colly/v2 v2.1.0       // Scrapping
    github.com/charmbracelet/log v0.3.1      // Logging
    github.com/google/uuid v1.5.0             // IDs
)
```

C'est tout! **Vraiment minimal** - pas de Gin (pas d'API), pas de base de données, pas de dépendances complexes.

---

## 8️⃣ **Étapes de Construction (Détaillées)**

### **Phase 1: Fondations (2-3h)**
1. ✅ Créer structure + go.mod
2. ✅ Models domaine (Review, Film, User, GameState)
3. ✅ Port ReviewProvider interface
4. ✅ Tests unitaires pour les models

### **Phase 2: Scraper Letterboxd (4-5h)**
5. ✅ Scrapper Gocolly basic
6. ✅ Parse una página de reviews
7. ✅ Pagination (jusqu'à fin)
8. ✅ Filtering (bonnes reviews seulement)
9. ✅ Tests du scrapper

### **Phase 3: Logique du Jeu (2-3h)**
10. ✅ GameService.LoadGame()
11. ✅ GameService.SubmitAnswer()
12. ✅ Scoring logic
13. ✅ Tests logique jeu

### **Phase 4: CLI Interactive (2-3h)**
14. ✅ Prompts (usernames, question count)
15. ✅ Boucle de jeu
16. ✅ Affichage questions/réponses avec colors
17. ✅ Écran final

### **Phase 5: Polish (1-2h)**
18. ✅ Error handling
19. ✅ Spinners de loading
20. ✅ README + guide
21. ✅ Taskfile (build/run)

**Total: ~1 semaine max**

---

## 🔟 **Décisions d'Architecture Clés**

| Décision | Raison |
|----------|--------|
| **CLI-first, pas API** | Jeu local/offline, pas besoin client-serveur |
| **Tout en mémoire** | Données dans GameState, pas de persistence |
| **Single service** | GameService assez simple pour pas de dependencies |
| **Gocolly scraper** | Même pattern que twin-pick, efficace |
| **Pas de cache** | One-time load, pas de réutilisation |
| **Minimaliste** | Juste le strict pour jouer, zéro bloat |

---

## 1️⃣1️⃣ **Flow d'Utilisation**

```
$ go run ./cmd/review-guess/main.go

=== Review Guesser ===

Quels pseudos Letterboxd? (comma-separated)
> alice,bob,charlie

Combien de questions?
> 10

⏳ Récupération des reviews...
alice: 47 reviews
bob: 128 reviews
charlie: 85 reviews

✅ 260 reviews chargées, 10 questions générées

=== Question 1/10 ===
« Watched it yesterday and WOW, the cinematography was absolutely stunning! »

Qui l'a écrite?
> bob

Quel film?
> godfather

✅ Correct! Film: The Godfather | Auteur: Bob
Points: +100
Score total: 100/100

Suivant? [Appuie]

...

=== FIN ===
Final Score: 750/1000 (75%)
```

---

## 1️⃣2️⃣ **Prochaine Étape**

→ **Commencer Phase 1 immédiatement** ou faire des ajustements?

Questions avant de coder:
- Règles de scoring: 100pt si les deux, 50pt chacun, ou autre?
- Affichage de la review: masquer seulement auteur/film, ou plus?
- Films avec plusieurs auteurs: match partiel compte?
- Interface: plus "MineCraft" ou "old terminal retro"?
