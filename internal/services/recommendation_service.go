package services

import (
	"movie-watchlist/internal/models"
	"movie-watchlist/internal/repositories"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecommendationService struct {
	movieRepo              *repositories.MovieRepository
	ratingRepo             *repositories.RatingRepository
	watchlistRepo          *repositories.WatchlistRepository
	recommendationRepo      *repositories.RecommendationRepository
}

func NewRecommendationService(movieRepo *repositories.MovieRepository, ratingRepo *repositories.RatingRepository, watchlistRepo *repositories.WatchlistRepository) *RecommendationService {
	return &RecommendationService{
		movieRepo:         movieRepo,
		ratingRepo:        ratingRepo,
		watchlistRepo:     watchlistRepo,
		recommendationRepo: repositories.NewRecommendationRepository(movieRepo.GetDB()),
	}
}

func (s *RecommendationService) GetRecommendations(userID primitive.ObjectID, limit int) ([]models.Movie, error) {
	// Step 1: Get user's preferred genres (rated 4+ stars)
	preferredGenres, err := s.recommendationRepo.GetHighRatedGenres(userID, 4)
	if err != nil {
		return nil, err
	}

	// Step 2: Get movies to exclude (already rated + in watchlist)
	excludeMovieIDs, err := s.recommendationRepo.GetMoviesToExclude(userID)
	if err != nil {
		return nil, err
	}

	// Step 3: Generate recommendations based on preferred genres
	recommendations := s.generateGenreBasedRecommendations(preferredGenres, excludeMovieIDs, limit)

	// Step 4: If not enough recommendations, add popular movies as fallback
	if len(recommendations) < limit {
		fallbackMovies := s.getFallbackRecommendations(excludeMovieIDs, limit-len(recommendations))
		recommendations = append(recommendations, fallbackMovies...)
	}

	// Step 5: Return limited results (deterministic ordering)
	return s.limitResults(recommendations, limit), nil
}

// getPreferredGenres identifies genres user rated 4+ stars
func (s *RecommendationService) getPreferredGenres(userID primitive.ObjectID) ([]string, error) {
	return s.recommendationRepo.GetHighRatedGenres(userID, 4)
}

// getExcludedMovieIDs returns IDs of movies already rated or in watchlist
func (s *RecommendationService) getExcludedMovieIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	return s.recommendationRepo.GetMoviesToExclude(userID)
}

// generateGenreBasedRecommendations creates recommendations from preferred genres
func (s *RecommendationService) generateGenreBasedRecommendations(preferredGenres []string, excludeMovieIDs []primitive.ObjectID, limit int) []models.Movie {
	var recommendations []models.Movie

	// Process each preferred genre in order
	for _, genre := range preferredGenres {
		if len(recommendations) >= limit {
			break
		}

		// Get movies in this genre, excluding already watched/rated movies
		movies, err := s.recommendationRepo.GetMoviesByGenreExcludingIDs(genre, excludeMovieIDs, limit-len(recommendations))
		if err != nil {
			continue
		}

		// Add movies (deterministic order by IMDb rating)
		for _, movie := range movies {
			if len(recommendations) >= limit {
				break
			}
			recommendations = append(recommendations, movie)
		}
	}

	return recommendations
}

// getFallbackRecommendations provides popular movies when genre-based recommendations are insufficient
func (s *RecommendationService) getFallbackRecommendations(excludeMovieIDs []primitive.ObjectID, limit int) []models.Movie {
	var fallback []models.Movie

	// Get all movies as fallback
	allMovies, err := s.movieRepo.FindAll()
	if err != nil {
		return fallback
	}

	// Create exclusion map for faster lookup
	excludeMap := make(map[primitive.ObjectID]bool)
	for _, id := range excludeMovieIDs {
		excludeMap[id] = true
	}

	// Add movies that aren't excluded (deterministic order by IMDb rating)
	for _, movie := range allMovies {
		if len(fallback) >= limit {
			break
		}
		if !excludeMap[movie.ID] {
			fallback = append(fallback, movie)
		}
	}

	return fallback
}

// limitResults returns a deterministic slice of results
func (s *RecommendationService) limitResults(movies []models.Movie, limit int) []models.Movie {
	if len(movies) <= limit {
		return movies
	}
	return movies[:limit]
}

func (s *RecommendationService) normalizeGenre(genre string) string {
	genre = strings.ToLower(strings.TrimSpace(genre))
	if strings.Contains(genre, ",") {
		parts := strings.Split(genre, ",")
		return strings.TrimSpace(parts[0])
	}
	return genre
}
