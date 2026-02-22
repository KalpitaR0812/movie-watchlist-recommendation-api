package handlers

import (
	"movie-watchlist/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovieHandler struct {
	movieService *services.MovieService
}

func NewMovieHandler(movieService *services.MovieService) *MovieHandler {
	return &MovieHandler{movieService: movieService}
}

func (h *MovieHandler) SearchMovies(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	movies, err := h.movieService.SearchMovies(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

func (h *MovieHandler) GetMovie(c *gin.Context) {
	idParam := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	movie, err := h.movieService.GetMovieByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

// GetMovieByIMDbID fetches movie details by IMDb ID
func (h *MovieHandler) GetMovieByIMDbID(c *gin.Context) {
	imdbID := c.Query("imdb_id")
	if imdbID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IMDb ID is required"})
		return
	}

	movie, err := h.movieService.GetOrCreateByIMDbID(imdbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}
