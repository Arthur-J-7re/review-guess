package domain

// Review représente une critique d'utilisateur sur un film
type Review struct {
	Author   string  `json:"author"`
	Title    string  `json:"title"`
	Slug     string  `json:"slug"`
	Content  string  `json:"content"`
	Rating   float64 `json:"rating"`
	Liked    bool    `json:"liked"`
	Spoilers bool    `json:"spoilers"`
}

// Reviews représente une collection de reviews d'un utilisateur
type Reviews struct {
	Count   int       `json:"count"`
	Reviews []*Review `json:"reviews"`
}
