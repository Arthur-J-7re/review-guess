package domain

// ===== SCRAPING =====

// ReviewProvider = Interface pour récupérer les reviews d'un utilisateur
type ReviewProvider interface {
	// FetchUserReviews récupère toutes les reviews d'un utilisateur (paginated)
	FetchUserReviews(username string) ([]*Review, error)
}

// ===== REPOSITORIES =====

// PlayerRepository gère les opérations sur les joueurs
type PlayerRepository interface {
	// Get récupère un joueur par ID
	Get(id string) (*Player, error)
	// Create crée un nouveau joueur
	Create(player *Player) error
	// Update met à jour un joueur
	Update(player *Player) error
	// List récupère tous les joueurs
	List() ([]*Player, error)
}

// LetterboxdReviewerRepository gère les opérations sur les reviewers Letterboxd
type LetterboxdReviewerRepository interface {
	// Get récupère un reviewer par ID
	Get(id string) (*LetterboxdReviewer, error)
	// GetByUsername récupère un reviewer par son username Letterboxd
	GetByUsername(username string) (*LetterboxdReviewer, error)
	// Create crée un nouveau reviewer
	Create(reviewer *LetterboxdReviewer) error
	// Update met à jour un reviewer
	Update(reviewer *LetterboxdReviewer) error
	// List récupère tous les reviewers
	List() ([]*LetterboxdReviewer, error)
}

// PlayerReviewerLinkRepository gère les liens entre joueurs et reviewers
type PlayerReviewerLinkRepository interface {
	// Create crée un lien entre un joueur et un reviewer
	Create(link *PlayerReviewerLink) error
	// Get récupère le lien entre un joueur et un reviewer
	Get(playerID, reviewerID string) (*PlayerReviewerLink, error)
	// GetPlayerReviewers récupère tous les reviewers d'un joueur
	GetPlayerReviewers(playerID string) ([]*LetterboxdReviewer, error)
	// Delete supprime le lien
	Delete(playerID, reviewerID string) error
}

// MovieRepository gère les opérations sur les films
type MovieRepository interface {
	// Get récupère un film par ID
	Get(id string) (*Movie, error)
	// GetByTmdbID récupère un film par son ID TMDB
	GetByTmdbID(tmdbID int) (*Movie, error)
	// GetByLetterboxdSlug récupère un film par son slug Letterboxd
	GetByLetterboxdSlug(slug string) (*Movie, error)
	// Create crée un nouveau film
	Create(movie *Movie) error
	// Update met à jour un film
	Update(movie *Movie) error
	// GetAll récupère tous les films
	GetAll() ([]*Movie, error)
	// GetMany récupère plusieurs films par leurs IDs
	GetMany(ids []string) ([]*Movie, error)
}

// ReviewRepository gère les opérations sur les critiques
type ReviewRepository interface {
	// Get récupère une review par ID
	Get(id string) (*Review, error)
	// GetByReviewerAndMovie récupère la review d'un reviewer pour un film
	GetByReviewerAndMovie(reviewerID, movieID string) (*Review, error)
	// GetUsableReviewsForReviewer récupère toutes les reviews utilisables d'un reviewer
	GetUsableReviewsForReviewer(reviewerID string) ([]*Review, error)
	// GetRandomUsableReview récupère une review aléatoire marquée comme usable
	GetRandomUsableReview() (*Review, error)
	// Create crée une nouvelle review
	Create(review *Review) error
	// Update met à jour une review
	Update(review *Review) error
	// CreateBatch crée plusieurs reviews en une seule transaction
	CreateBatch(reviews []*Review) error
}

// MovieSimilarityRepository gère les relations de similarité entre films
type MovieSimilarityRepository interface {
	// Get récupère la similarité entre deux films
	Get(movieAID, movieBID string) (*MovieSimilarity, error)
	// GetTopSimilarMovies récupère les N films les plus similaires à un film donné
	GetTopSimilarMovies(movieID string, limit int) ([]*MovieSimilarity, error)
	// Create crée une nouvelle relation de similarité
	Create(similarity *MovieSimilarity) error
	// CreateBatch crée plusieurs relations en une seule transaction
	CreateBatch(similarities []*MovieSimilarity) error
	// DeleteByMovie supprime tous les enregistrements de similarité pour un film
	DeleteByMovie(movieID string) error
}

// ReviewerMovieRepository gère le suivi des films regardés par les reviewers
type ReviewerMovieRepository interface {
	// Get récupère l'enregistrement d'un film regardé par un reviewer
	Get(reviewerID, movieID string) (*ReviewerMovie, error)
	// GetMoviesWatchedByReviewer récupère tous les films regardés par un reviewer
	GetMoviesWatchedByReviewer(reviewerID string) ([]*ReviewerMovie, error)
	// GetMoviesNotWatchedByReviewer récupère les films NON regardés par un reviewer (pour leurres)
	GetMoviesNotWatchedByReviewer(reviewerID string, limit int) ([]*Movie, error)
	// Create crée un nouvel enregistrement
	Create(reviewerMovie *ReviewerMovie) error
	// Update met à jour un enregistrement
	Update(reviewerMovie *ReviewerMovie) error
	// CreateBatch crée plusieurs enregistrements
	CreateBatch(reviewerMovies []*ReviewerMovie) error
}

// PersonRepository gère les acteurs et réalisateurs
type PersonRepository interface {
	// Get récupère une personne par ID
	Get(id string) (*Person, error)
	// GetByTmdbID récupère une personne par son ID TMDB
	GetByTmdbID(tmdbID int) (*Person, error)
	// Create crée une nouvelle personne
	Create(person *Person) error
	// GetMovieCast récupère le casting d'un film
	GetMovieCast(movieID string) ([]*Person, error)
	// GetMovieDirectors récupère les réalisateurs d'un film
	GetMovieDirectors(movieID string) ([]*Person, error)
}

// QuizHistoryRepository gère l'historique des réponses au quiz
type QuizHistoryRepository interface {
	// Create crée une nouvelle entrée d'historique
	Create(answer *QuizAnswer) error
	// GetPlayerHistory récupère l'historique total d'un joueur
	GetPlayerHistory(playerID string) ([]*QuizAnswer, error)
	// GetPlayerScores calcule les scores du joueur
	GetPlayerScores(playerID string) (totalAnswered int, correctAnswers int, err error)
}
