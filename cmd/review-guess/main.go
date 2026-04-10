package main

import (
	"log"
	"net/http"
	"os"

	"review-guess/internal/adapters/httpapi"
	"review-guess/internal/infrastructure/database"
	"review-guess/internal/infrastructure/scrapper"
)

func main() {
	// === DATABASE ===
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./review-guess.db"
	}
	db, err := database.NewSQLiteClient(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// === REPOSITORIES ===
	sqlDB := db.GetDB()

	playerRepo := database.NewPlayerRepository(sqlDB)
	reviewerRepo := database.NewLetterboxdReviewerRepository(sqlDB)
	linkRepo := database.NewPlayerReviewerLinkRepository(sqlDB)

	movieRepo := database.NewMovieRepository(sqlDB)
	reviewRepo := database.NewReviewRepository(sqlDB)
	reviewerMovieRepo := database.NewReviewerMovieRepository(sqlDB)

	similarityRepo := database.NewMovieSimilarityRepository(sqlDB)
	quizRepo := database.NewQuizHistoryRepository(sqlDB)
	personRepo := database.NewPersonRepository(sqlDB)

	// === SCRAPER ===
	scraper := scrapper.NewScrapper()

	// === ROUTER ===
	router := httpapi.NewRouter(
		playerRepo, reviewerRepo, linkRepo,
		movieRepo, reviewRepo, reviewerMovieRepo,
		similarityRepo, quizRepo, personRepo,
		scraper,
	)
	mux := router.Register()

	// === START SERVER ===
	port := ":8080"
	log.Println("🎬 Review-Guess server starting on " + port)
	log.Printf("📁 Database: %s\n", dbPath)
	log.Printf("🔗 API: http://localhost%s\n", port)
	log.Println()
	log.Println("📚 Available endpoints:")
	log.Println()
	log.Println("  Health:")
	log.Println("    GET  /health")
	log.Println()
	log.Println("  Scraper (Letterboxd reviews):")
	log.Println("    GET  /api/reviews?username=alice,bob")
	log.Println()
	log.Println("  Reviewers (data sources):")
	log.Println("    GET  /api/reviewers")
	log.Println("    GET  /api/reviewers/{username}")
	log.Println()
	log.Println("  Quiz (game endpoints):")
	log.Println("    GET  /api/quiz/next?player_id=xxx")
	log.Println("    POST /api/quiz/answer (form: player_id, review_id, answer_id)")
	log.Println("    GET  /api/quiz/stats?player_id=xxx")
	log.Println()
	log.Println("📖 Full documentation: see API.md")
	log.Println()

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
