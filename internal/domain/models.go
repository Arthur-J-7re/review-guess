package domain

import "time"

// ===== PLAYERS (Quiz Players) =====

// Player représente un joueur du quiz
// Peut être anonyme (is_logged_in=false) ou connecté
type Player struct {
	ID                     string     `json:"id"`
	Nickname               string     `json:"nickname,omitempty"`
	IsLoggedIn             bool       `json:"is_logged_in"`
	LastReviewedReviewerID *string    `json:"last_reviewed_reviewer_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	LastPlayedAt           *time.Time `json:"last_played_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// ===== LETTERBOXD REVIEWERS =====

// LetterboxdReviewer représente un compte Letterboxd qu'on scrape
// Indépendant des joueurs - un reviewer n'a pas besoin de jouer au quiz
type LetterboxdReviewer struct {
	ID                     string     `json:"id"`
	LetterboxdUsername     string     `json:"letterboxd_username"`
	TotalReviews           int        `json:"total_reviews"`
	TotalMoviesWatched     int        `json:"total_movies_watched"`
	LastReviewPageScrapped int        `json:"last_review_page_scrapped"`
	LastMoviePageScrapped  int        `json:"last_movie_page_scrapped"`
	LastScrappedAt         *time.Time `json:"last_scrapped_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// PlayerReviewerLink représente le lien entre un joueur et ses reviewers Letterboxd
type PlayerReviewerLink struct {
	PlayerID   string    `json:"player_id"`
	ReviewerID string    `json:"reviewer_id"`
	IsPrimary  bool      `json:"is_primary"`
	LinkedAt   time.Time `json:"linked_at"`
}

// ===== MOVIE =====

// Movie représente un film enrichi avec les données TMDB
type Movie struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Year           *int      `json:"year"`
	Description    string    `json:"description"`
	PosterURL      string    `json:"poster_url"`
	TmdbID         *int      `json:"tmdb_id"`
	LetterboxdSlug string    `json:"letterboxd_slug"`
	Directors      []*Person `json:"directors,omitempty"`
	Cast           []*Person `json:"cast,omitempty"`
	Genres         []string  `json:"genres,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Person représente un acteur ou réalisateur
type Person struct {
	ID             string    `json:"id"`
	TmdbID         *int      `json:"tmdb_id"`
	Name           string    `json:"name"`
	ProfilePicture string    `json:"profile_picture"`
	CharacterName  string    `json:"character_name,omitempty"` // Pour les acteurs
	RoleOrder      *int      `json:"role_order,omitempty"`     // Pour les acteurs (importance)
	Job            string    `json:"job,omitempty"`            // Pour les crews (Director, Producer, etc.)
	CreatedAt      time.Time `json:"created_at"`
}

// ===== REVIEW =====

// Review représente une critique d'utilisateur sur un film
type Review struct {
	ID         string    `json:"id"`
	ReviewerID string    `json:"reviewer_id"`
	MovieID    string    `json:"movie_id"`
	Author     string    `json:"author,omitempty"` // Letterboxd username
	Title      string    `json:"title"`
	Slug       string    `json:"slug,omitempty"`
	Content    string    `json:"content"`
	Rating     float64   `json:"rating"`
	Liked      bool      `json:"liked"`
	Spoilers   bool      `json:"spoilers"`
	Usable     bool      `json:"usable"` // Si la review peut être utilisée dans les quiz
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Reviews représente une collection de reviews d'un utilisateur
type Reviews struct {
	Count   int       `json:"count"`
	Reviews []*Review `json:"reviews"`
}

// ===== MOVIE SIMILARITY =====

// MovieSimilarity représente le score de similarité entre deux films
type MovieSimilarity struct {
	MovieAID        string    `json:"movie_a_id"`
	MovieBID        string    `json:"movie_b_id"`
	SimilarityScore float64   `json:"similarity_score"`
	SharedDirectors int       `json:"shared_directors"`
	SharedActors    int       `json:"shared_actors"`
	SharedGenres    int       `json:"shared_genres"`
	YearProximity   *int      `json:"year_proximity"`
	CalculatedAt    time.Time `json:"calculated_at"`
}

// ===== USER MOVIE =====

// ReviewerMovie représente un film regardé par un reviewer (avec ou sans review)
type ReviewerMovie struct {
	ReviewerID string     `json:"reviewer_id"`
	MovieID    string     `json:"movie_id"`
	HasReview  bool       `json:"has_review"`
	Rating     *float64   `json:"rating"`
	WatchedAt  *time.Time `json:"watched_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ===== QUIZ =====

// QuizQuestion représente une question de quiz
type QuizQuestion struct {
	ReviewID     string   `json:"review_id"`
	ReviewText   string   `json:"review_text"`
	CorrectMovie *Movie   `json:"correct_movie"`
	Options      []*Movie `json:"options"`
}

// QuizAnswer représente une réponse utilisateur
type QuizAnswer struct {
	ID             string    `json:"id"`
	PlayerID       string    `json:"player_id"`
	ReviewID       string    `json:"review_id"`
	CorrectMovieID string    `json:"correct_movie_id"`
	PlayerAnswerID *string   `json:"player_answer_id"`
	IsCorrect      bool      `json:"is_correct"`
	Options        []string  `json:"options"` // IDs des films proposés
	AnsweredAt     time.Time `json:"answered_at"`
}
