package domain

// ReviewProvider = Interface pour récupérer les reviews d'un utilisateur
type ReviewProvider interface {
	// FetchUserReviews récupère toutes les reviews d'un utilisateur (paginated)
	FetchUserReviews(username string) ([]*Review, error)
}
