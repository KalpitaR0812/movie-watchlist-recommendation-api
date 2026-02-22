package main

import (
	"log"
	"movie-watchlist/internal/config"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/handlers"
	"movie-watchlist/internal/middleware"
	"movie-watchlist/internal/repositories"
	"movie-watchlist/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	cfg := config.Load()

	// Validate required configuration
	if cfg.OMDbAPIKey == "" {
		log.Fatal("OMDb API key not configured. Please set OMDB_API_KEY in .env file or environment variables")
	}

	log.Println("Configuration loaded successfully")
	log.Printf("Database URL: %s", cfg.DatabaseURL)
	log.Println("OMDb API key: configured")

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	userRepo := repositories.NewUserRepository(db)
	movieRepo := repositories.NewMovieRepository(db, cfg.OMDbAPIKey)
	watchlistRepo := repositories.NewWatchlistRepository(db)
	ratingRepo := repositories.NewRatingRepository(db)

	userService := services.NewUserService(userRepo)
	movieService := services.NewMovieService(movieRepo, cfg.OMDbAPIKey)
	watchlistService := services.NewWatchlistService(watchlistRepo)
	ratingService := services.NewRatingService(ratingRepo)
	recommendationService := services.NewRecommendationService(movieRepo, ratingRepo, watchlistRepo)

	authHandler := handlers.NewAuthHandler(userService, cfg.JWTSecret)
	movieHandler := handlers.NewMovieHandler(movieService)
	watchlistHandler := handlers.NewWatchlistHandler(watchlistService)
	ratingHandler := handlers.NewRatingHandler(ratingService)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationService)

	r := gin.Default()

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		api.GET("/movies/search", movieHandler.SearchMovies)
		api.GET("/movies/:id", movieHandler.GetMovie)
		api.GET("/movies/by-imdb", movieHandler.GetMovieByIMDbID)
		api.POST("/watchlist", watchlistHandler.AddToWatchlist)
		api.DELETE("/watchlist/:movieId", watchlistHandler.RemoveFromWatchlist)
		api.GET("/watchlist", watchlistHandler.GetWatchlist)
		api.POST("/ratings", ratingHandler.RateMovie)
		api.PUT("/ratings/:movieId", ratingHandler.UpdateRating)
		api.GET("/ratings", ratingHandler.GetUserRatings)
		api.GET("/recommendations", recommendationHandler.GetRecommendations)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
