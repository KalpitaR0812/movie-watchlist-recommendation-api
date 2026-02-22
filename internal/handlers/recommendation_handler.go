package handlers

import (
	"movie-watchlist/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RecommendationHandler struct {
	recommendationService *services.RecommendationService
}

func NewRecommendationHandler(recommendationService *services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService}
}

func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(primitive.ObjectID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	limit := 10 // Default limit
	recommendations, err := h.recommendationService.GetRecommendations(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response with additional metadata
	var formattedRecommendations []gin.H
	for _, movie := range recommendations {
		formattedRecommendations = append(formattedRecommendations, gin.H{
			"id":          movie.ID,
			"title":       movie.Title,
			"year":        movie.Year,
			"genre":       movie.Genre,
			"director":    movie.Director,
			"poster":      movie.Poster,
			"imdb_rating": movie.IMDbRating,
			"imdb_id":     movie.IMDbID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": formattedRecommendations,
		"count":         len(formattedRecommendations),
		"limit":         limit,
		"algorithm":     "rule-based",
		"criteria":      "Genres rated 4+ stars, excluding rated and watchlist movies",
	})
}
