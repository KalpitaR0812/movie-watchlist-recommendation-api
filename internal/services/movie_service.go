package services

import (
	"context"
	"encoding/json"
	"fmt"
	"movie-watchlist/internal/models"
	"movie-watchlist/internal/repositories"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type OMDbSearchResponse struct {
	Search       []OMDbResponse `json:"Search"`
	TotalResults string          `json:"totalResults"`
	Response     string          `json:"Response"`
	Error        string          `json:"Error"`
}

type MovieService struct {
	movieRepo *repositories.MovieRepository
	apiKey    string
	client    *http.Client
}

func NewMovieService(movieRepo *repositories.MovieRepository, apiKey string) *MovieService {
	return &MovieService{
		movieRepo: movieRepo,
		apiKey:    apiKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *MovieService) SearchMovies(ctx context.Context, query string) ([]OMDbResponse, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("OMDb API key not configured")
	}

	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// URL encode the query for safe HTTP requests
	encodedQuery := url.QueryEscape(query)
	requestURL := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&s=%s", s.apiKey, encodedQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OMDb API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OMDb API returned status code: %d", resp.StatusCode)
	}

	var searchResp OMDbSearchResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode OMDb API response: %w", err)
	}

	// Check for API-level errors
	if searchResp.Response == "False" {
		if searchResp.Error != "" {
			return nil, fmt.Errorf("OMDb API error: %s", searchResp.Error)
		}
		return nil, fmt.Errorf("OMDb API returned an error response")
	}

	if len(searchResp.Search) == 0 {
		return []OMDbResponse{}, nil
	}

	// Cache full movie details for each search result
	for _, item := range searchResp.Search {
		// 1. Check if movie already exists
		existing, _ := s.movieRepo.FindByIMDbID(item.IMDbID)
		if existing != nil {
			continue
		}

		// 2. Fetch FULL movie details
		details, err := s.fetchMovieDetails(ctx, item.IMDbID)
		if err != nil {
			continue
		}

		// 3. Save FULL movie (genre INCLUDED)
		movie := &models.Movie{
			IMDbID:     details.IMDbID,
			Title:      strings.TrimSpace(details.Title),
			Year:       strings.TrimSpace(details.Year),
			Genre:      strings.TrimSpace(details.Genre),        // THIS WAS MISSING
			Director:   strings.TrimSpace(details.Director),
			Plot:       strings.TrimSpace(details.Plot),
			Poster:     strings.TrimSpace(details.Poster),
			Runtime:    strings.TrimSpace(details.Runtime),
			IMDbRating: strings.TrimSpace(details.IMDbRating),
			CachedAt:   time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		_ = s.movieRepo.Create(movie)
	}

	return searchResp.Search, nil
}

// Helper method to fetch movie details by IMDb ID
func (s *MovieService) fetchMovieDetails(ctx context.Context, imdbID string) (*OMDbResponse, error) {
	// URL encode the IMDb ID for safe HTTP requests
	encodedIMDbID := url.QueryEscape(imdbID)
	requestURL := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%s", s.apiKey, encodedIMDbID)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
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

	return &omdbResp, nil
}

func (s *MovieService) GetMovieDetails(ctx context.Context, imdbID string) (*models.Movie, error) {
	// Validate IMDb ID format
	if strings.TrimSpace(imdbID) == "" {
		return nil, fmt.Errorf("IMDb ID cannot be empty")
	}

	// Check cache first
	if movie, err := s.movieRepo.FindByIMDbID(imdbID); err == nil {
		return movie, nil
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("OMDb API key not configured")
	}

	// URL encode the IMDb ID for safe HTTP requests
	encodedIMDbID := url.QueryEscape(imdbID)
	requestURL := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%s", s.apiKey, encodedIMDbID)

	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
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

	movie := &models.Movie{
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
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.movieRepo.Create(movie); err != nil {
		return nil, fmt.Errorf("failed to cache movie data: %w", err)
	}

	return movie, nil
}

func (s *MovieService) GetMovieByID(id primitive.ObjectID) (*models.Movie, error) {
	return s.movieRepo.FindByID(id)
}

// GetOrCreateByIMDbID fetches movie by IMDb ID, creating from OMDb if not found
func (s *MovieService) GetOrCreateByIMDbID(imdbID string) (*models.Movie, error) {
	movie, err := s.movieRepo.GetOrCreateByIMDbID(imdbID)
	if err != nil {
		return nil, err
	}
	return movie, nil
}
