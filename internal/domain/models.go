package domain

// User représente un utilisateur Letterboxd
type User struct {
	Username string    // ID unique
	Name     string    // nom d'affichage (optionnel)
	Reviews  []*Review // reviews fetchées au démarrage
}

// Film représente un film sur Letterboxd
type Film struct {
	Slug      string // film slug (unique)
	Title     string
	Year      int
	Directors []string // Pour matching multi-réalisateurs
	Poster    string   // URL poster (optionnel pour affichage)
}

// Review = une critique d'utilisateur sur un film
type Review struct {
	ID       string // unique (username-filmslug)
	Author   string // username
	Film     *Film  // référence au film
	Content  string // le texte de la review
	Rating   int    // note 0-5 (0 = "watched" sans note, 1-5 = ★)
	Liked    bool   // ❤️ ou pas
	Spoilers bool   // marquée comme spoilers?
}

// Question = une question pendant le jeu
type Question struct {
	ReviewIndex int     // idx dans la liste des reviews chargées
	Review      *Review // la review en question
	Difficulty  float32 // calculé au runtime
}

// GameState = état du jeu en cours (en mémoire seulement)
type GameState struct {
	Users      []*User     // tous les utilisateurs loadés
	AllReviews []*Review   // fusion de toutes les reviews
	Questions  []*Question // les questions de ce jeu
	Answers    []*Answer   // réponses du joueur
	Score      int
	CurrentIdx int // index question actuelle
}

// Answer = réponse du joueur à une question
type Answer struct {
	QuestionIdx   int
	GuessedUser   string // username deviné
	GuessedFilm   string // film slug deviné
	IsCorrectUser bool
	IsCorrectFilm bool
	Points        int
}

// GameResults = résultats finaux
type GameResults struct {
	TotalScore  int
	TotalPoints int
	Percentage  float32
	Answers     []*Answer
}
