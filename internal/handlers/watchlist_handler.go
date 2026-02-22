package handlers

import (
	"movie-watchlist/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WatchlistHandler struct {
	watchlistService *services.WatchlistService
}

func NewWatchlistHandler(watchlistService *services.WatchlistService) *WatchlistHandler {
	return &WatchlistHandler{watchlistService: watchlistService}
}

type AddToWatchlistRequest struct {
	MovieID string `json:"movie_id" binding:"required"`
}

func (h *WatchlistHandler) AddToWatchlist(c *gin.Context) {
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

	var req AddToWatchlistRequest
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

	err = h.watchlistService.AddToWatchlist(userID, movieID)
	if err != nil {
		if err.Error() == "movie already in watchlist" {
			c.JSON(http.StatusConflict, gin.H{"error": "Movie is already in your watchlist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Movie added to watchlist successfully",
		"movie_id": req.MovieID,
	})
}

func (h *WatchlistHandler) RemoveFromWatchlist(c *gin.Context) {
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

	err = h.watchlistService.RemoveFromWatchlist(userID, movieID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Movie removed from watchlist successfully",
		"movie_id": movieIDParam,
	})
}

func (h *WatchlistHandler) GetWatchlist(c *gin.Context) {
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

	watchlist, err := h.watchlistService.GetUserWatchlist(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format response with movie details
	var watchlistResponse []gin.H
	for _, item := range watchlist {
		watchlistResponse = append(watchlistResponse, gin.H{
			"id":        item.ID,
			"added_at":  item.AddedAt,
			"movie_id":  item.MovieID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"watchlist": watchlistResponse,
		"count":     len(watchlistResponse),
	})
}
