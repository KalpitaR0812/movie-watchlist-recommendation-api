package services

import (
	"errors"
	"movie-watchlist/internal/models"
	"movie-watchlist/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingService struct {
	ratingRepo *repositories.RatingRepository
}

func NewRatingService(ratingRepo *repositories.RatingRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo}
}

func (s *RatingService) RateMovie(userID primitive.ObjectID, movieID primitive.ObjectID, rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5 stars")
	}

	// Check if user has already rated this movie
	existing, err := s.ratingRepo.GetUserRating(userID, movieID)
	if err == nil && existing != nil {
		return errors.New("user has already rated this movie")
	}

	newRating := &models.Rating{
		UserID:  userID,
		MovieID: movieID,
		Rating:  rating,
	}

	return s.ratingRepo.Create(newRating)
}

func (s *RatingService) UpdateRating(userID primitive.ObjectID, movieID primitive.ObjectID, rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5 stars")
	}

	// Check if rating exists before updating
	existing, err := s.ratingRepo.GetUserRating(userID, movieID)
	if err != nil {
		return errors.New("rating not found")
	}

	if existing == nil {
		return errors.New("rating not found")
	}

	return s.ratingRepo.Update(userID, movieID, rating)
}

func (s *RatingService) GetUserRatings(userID primitive.ObjectID) ([]models.Rating, error) {
	return s.ratingRepo.GetUserRatings(userID)
}

func (s *RatingService) GetUserRating(userID primitive.ObjectID, movieID primitive.ObjectID) (*models.Rating, error) {
	return s.ratingRepo.GetUserRating(userID, movieID)
}
