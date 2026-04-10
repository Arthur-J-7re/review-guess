-- ===== PLAYERS (Quiz Players) =====
-- Players can be anonymous (guest) or logged in
-- They play the quiz and their scores are tracked
CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    nickname TEXT,                        -- Optional: nickname for leaderboards
    is_logged_in BOOLEAN DEFAULT FALSE,
    last_reviewed_reviewer_id TEXT,      -- Last reviewer they played from
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_played_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===== LETTERBOXD REVIEWERS (Review Sources) =====
-- Letterboxd accounts we scrape reviews from
-- Independent from players - reviewers don't need to play the quiz
CREATE TABLE IF NOT EXISTS letterboxd_reviewers (
    id TEXT PRIMARY KEY,
    letterboxd_username TEXT UNIQUE NOT NULL,
    total_reviews INT DEFAULT 0,
    total_movies_watched INT DEFAULT 0,
    last_review_page_scrapped INT DEFAULT 0,
    last_movie_page_scrapped INT DEFAULT 0,
    last_scrapped_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===== PLAYER-REVIEWER LINK (Optional) =====
-- Links a logged-in player to their Letterboxd reviewer account
-- Allows players to play with their own reviews or track multiple reviewers
CREATE TABLE IF NOT EXISTS player_reviewer_links (
    player_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES letterboxd_reviewers(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE,   -- Primary reviewer for this player
    linked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (player_id, reviewer_id)
);

-- Movies (enriched with TMDB data)
CREATE TABLE IF NOT EXISTS movies (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    year INTEGER,
    description TEXT,
    poster_url TEXT,
    tmdb_id INTEGER UNIQUE,
    letterboxd_slug TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- People (Actors & Directors from TMDB)
CREATE TABLE IF NOT EXISTS people (
    id TEXT PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    name TEXT NOT NULL,
    profile_picture TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Movie Cast (Actors)
CREATE TABLE IF NOT EXISTS movie_cast (
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    character_name TEXT,
    role_order INTEGER,
    PRIMARY KEY (movie_id, person_id)
);

-- Movie Crew (Directors, Producers, etc.)
CREATE TABLE IF NOT EXISTS movie_crew (
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    person_id TEXT NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    job TEXT NOT NULL,
    PRIMARY KEY (movie_id, person_id, job)
);

-- Genres (from TMDB)
CREATE TABLE IF NOT EXISTS genres (
    id INT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Movie Genres
CREATE TABLE IF NOT EXISTS movie_genres (
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    genre_id INT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, genre_id)
);

-- Reviews (from Letterboxd scraping)
CREATE TABLE IF NOT EXISTS reviews (
    id TEXT PRIMARY KEY,
    reviewer_id TEXT NOT NULL REFERENCES letterboxd_reviewers(id) ON DELETE CASCADE,
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    title TEXT,
    content TEXT,
    rating FLOAT,
    liked BOOLEAN DEFAULT FALSE,
    spoilers BOOLEAN DEFAULT FALSE,
    usable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(reviewer_id, movie_id)
);

-- Reviewer Movies (tracking which movies each reviewer watched, with or without review)
CREATE TABLE IF NOT EXISTS reviewer_movies (
    reviewer_id TEXT NOT NULL REFERENCES letterboxd_reviewers(id) ON DELETE CASCADE,
    movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    has_review BOOLEAN DEFAULT FALSE,
    rating FLOAT,
    watched_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (reviewer_id, movie_id)
);

-- Movie Similarities (pre-calculated relations between movies)
CREATE TABLE IF NOT EXISTS movie_similarities (
    movie_a TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    movie_b TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    similarity_score FLOAT NOT NULL,
    shared_directors INT DEFAULT 0,
    shared_actors INT DEFAULT 0,
    shared_genres INT DEFAULT 0,
    year_proximity INT,
    calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (movie_a, movie_b)
);

-- Quiz History (track which questions players answered)
CREATE TABLE IF NOT EXISTS quiz_history (
    id TEXT PRIMARY KEY,
    player_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    review_id TEXT NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    correct_movie_id TEXT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    player_answer_id TEXT REFERENCES movies(id) ON DELETE SET NULL,
    is_correct BOOLEAN,
    options TEXT NOT NULL,  -- JSON array of movie IDs
    answered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_reviews_reviewer_id ON reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_reviews_movie_id ON reviews(movie_id);
CREATE INDEX IF NOT EXISTS idx_reviews_usable ON reviews(usable);
CREATE INDEX IF NOT EXISTS idx_reviewer_movies_reviewer_id ON reviewer_movies(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_reviewer_movies_movie_id ON reviewer_movies(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_cast_movie_id ON movie_cast(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_crew_movie_id ON movie_crew(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_genres_movie_id ON movie_genres(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_similarities_movie_a ON movie_similarities(movie_a);
CREATE INDEX IF NOT EXISTS idx_movie_similarities_score ON movie_similarities(similarity_score DESC);
CREATE INDEX IF NOT EXISTS idx_quiz_history_player_id ON quiz_history(player_id);
CREATE INDEX IF NOT EXISTS idx_quiz_history_review_id ON quiz_history(review_id);
CREATE INDEX IF NOT EXISTS idx_players_created_at ON players(created_at);
CREATE INDEX IF NOT EXISTS idx_players_last_played_at ON players(last_played_at);
CREATE INDEX IF NOT EXISTS idx_letterboxd_reviewers_username ON letterboxd_reviewers(letterboxd_username);
CREATE INDEX IF NOT EXISTS idx_player_reviewer_links_reviewer_id ON player_reviewer_links(reviewer_id);
