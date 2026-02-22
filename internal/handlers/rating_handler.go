package handlers

import (
	"movie-watchlist/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingHandler struct {
	ratingService *services.RatingService
}

func NewRatingHandler(ratingService *services.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

type RateMovieRequest struct {
	MovieID string `json:"movie_id" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
}

type UpdateRatingRequest struct {
	Rating int `json:"rating" binding:"required,min=1,max=5"`
}

func (h *RatingHandler) RateMovie(c *gin.Context) {
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

	var req RateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse movie ID from string to ObjectID
	movieID, err := primitive.ObjectIDFromHex(req.MovieID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID format"})
		return
	}

	err = h.ratingService.RateMovie(userID, movieID, req.Rating)
	if err != nil {
		if err.Error() == "user has already rated this movie" {
			c.JSON(http.StatusConflict, gin.H{"error": "You have already rated this movie. Use the update endpoint to change your rating."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Movie rated successfully",
		"movie_id": req.MovieID,
		"rating":   req.Rating,
		"stars":   h.getStarDisplay(req.Rating),
	})
}

func (h *RatingHandler) UpdateRating(c *gin.Context) {
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

	movieIDParam := c.Param("movieId")
	movieID, err := primitive.ObjectIDFromHex(movieIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID format"})
		return
	}

	var req UpdateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.ratingService.UpdateRating(userID, movieID, req.Rating)
	if err != nil {
		if err.Error() == "rating not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "You haven't rated this movie yet. Use the rate endpoint to add a rating."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rating updated successfully",
		"movie_id": movieIDParam,
		"rating":   req.Rating,
		"stars":   h.getStarDisplay(req.Rating),
	})
}

func (h *RatingHandler) GetUserRatings(c *gin.Context) {
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

	ratings, err := h.ratingService.GetUserRatings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response with star display
	var ratingsResponse []gin.H
	for _, rating := range ratings {
		ratingsResponse = append(ratingsResponse, gin.H{
			"id":         rating.ID,
			"movie_id":   rating.MovieID,
			"rating":     rating.Rating,
			"stars":      h.getStarDisplay(rating.Rating),
			"created_at": rating.CreatedAt,
			"updated_at": rating.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"ratings": ratingsResponse,
		"count":   len(ratingsResponse),
	})
}

// Helper function to convert rating to star display
func (h *RatingHandler) getStarDisplay(rating int) string {
	stars := ""
	for i := 1; i <= 5; i++ {
		if i <= rating {
			stars += "★"
		} else {
			stars += "☆"
		}
	}
	return stars
}
