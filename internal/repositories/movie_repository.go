package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/models"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MovieRepository struct {
	db     *database.MongoDB
	apiKey string
	client *http.Client
}

type OMDbResponse struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	IMDbID     string `json:"imdbID"`
	Genre      string `json:"Genre"`
	Director   string `json:"Director"`
	Plot       string `json:"Plot"`
	Poster     string `json:"Poster"`
	Runtime    string `json:"Runtime"`
	IMDbRating string `json:"imdbRating"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

func NewMovieRepository(db *database.MongoDB, apiKey string) *MovieRepository {
	return &MovieRepository{
		db:     db,
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *MovieRepository) Create(movie *models.Movie) error {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	movie.CreatedAt = getCurrentTime()
	movie.UpdatedAt = getCurrentTime()
	movie.CachedAt = time.Now()
	
	// Only set ID if it's empty (zero value)
	if movie.ID.IsZero() {
		movie.ID = primitive.NewObjectID()
	}
	
	result, err := collection.InsertOne(ctx, movie)
	if err != nil {
		return err
	}
	
	// Set the ID from the insertion result if not already set
	if movie.ID.IsZero() {
		movie.ID = result.InsertedID.(primitive.ObjectID)
	}
	return nil
}

func (r *MovieRepository) FindByID(id primitive.ObjectID) (*models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	var movie models.Movie
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepository) FindByIMDbID(imdbID string) (*models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	var movie models.Movie
	err := collection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &movie, nil
}

func (r *MovieRepository) FindByGenre(genre string) ([]models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	cursor, err := collection.Find(ctx, bson.M{"genre": bson.M{"$regex": genre, "$options": "i"}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepository) FindAll() ([]models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var movies []models.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *MovieRepository) GetOrCreateByIMDbID(imdbID string) (*models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	var movie models.Movie

	// 1. Try to find existing movie
	err := collection.FindOne(ctx, bson.M{"imdb_id": imdbID}).Decode(&movie)
	if err == nil {
		return &movie, nil
	}

	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// 2. Fetch full movie details from OMDb using i= endpoint
	if r.apiKey == "" {
		return nil, fmt.Errorf("OMDb API key not configured")
	}

	// URL encode the IMDb ID for safe HTTP requests
	encodedIMDbID := url.QueryEscape(imdbID)
	requestURL := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%s", r.apiKey, encodedIMDbID)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OMDb API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OMDb API returned status code: %d", resp.StatusCode)
	}

	var omdbResp OMDbResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&omdbResp); err != nil {
		return nil, fmt.Errorf("failed to decode OMDb API response: %w", err)
	}

	// Check for API-level errors
	if omdbResp.Response == "False" {
		if omdbResp.Error != "" {
			return nil, fmt.Errorf("OMDb API error: %s", omdbResp.Error)
		}
		return nil, fmt.Errorf("OMDb API returned an error response")
	}

	// Validate required fields
	if omdbResp.IMDbID == "" {
		return nil, fmt.Errorf("invalid movie data: missing IMDb ID")
	}
	if omdbResp.Title == "" {
		return nil, fmt.Errorf("invalid movie data: missing title")
	}
	if omdbResp.Genre == "" {
		return nil, fmt.Errorf("invalid movie data: missing genre")
	}

	// 3. Construct MongoDB movie with full details
	movie = models.Movie{
		ID:         primitive.NewObjectID(),
		IMDbID:     omdbResp.IMDbID,
		Title:      strings.TrimSpace(omdbResp.Title),
		Year:       strings.TrimSpace(omdbResp.Year),
		Genre:      strings.TrimSpace(omdbResp.Genre),
		Director:   strings.TrimSpace(omdbResp.Director),
		Plot:       strings.TrimSpace(omdbResp.Plot),
		Poster:     strings.TrimSpace(omdbResp.Poster),
		Runtime:    strings.TrimSpace(omdbResp.Runtime),
		IMDbRating: strings.TrimSpace(omdbResp.IMDbRating),
		CachedAt:   time.Now(),
		CreatedAt:  getCurrentTime(),
		UpdatedAt:  getCurrentTime(),
	}

	// 4. Insert into MongoDB
	_, err = collection.InsertOne(ctx, movie)
	if err != nil {
		return nil, fmt.Errorf("failed to cache movie data: %w", err)
	}

	// 5. RETURN THE MOVIE
	return &movie, nil
}

// GetDB returns the underlying MongoDB database instance
func (r *MovieRepository) GetDB() *database.MongoDB {
	return r.db
}
