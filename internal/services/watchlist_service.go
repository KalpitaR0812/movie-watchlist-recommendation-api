package services

import (
	"errors"
	"movie-watchlist/internal/models"
	"movie-watchlist/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WatchlistService struct {
	watchlistRepo *repositories.WatchlistRepository
}

func NewWatchlistService(watchlistRepo *repositories.WatchlistRepository) *WatchlistService {
	return &WatchlistService{watchlistRepo: watchlistRepo}
}

func (s *WatchlistService) AddToWatchlist(userID primitive.ObjectID, movieID primitive.ObjectID) error {
	exists, err := s.watchlistRepo.Exists(userID, movieID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("movie already in watchlist")
	}

	watchlist := &models.Watchlist{
		UserID:  userID,
		MovieID: movieID,
	}

	return s.watchlistRepo.Add(watchlist)
}

func (s *WatchlistService) RemoveFromWatchlist(userID primitive.ObjectID, movieID primitive.ObjectID) error {
	return s.watchlistRepo.Remove(userID, movieID)
}

func (s *WatchlistService) GetUserWatchlist(userID primitive.ObjectID) ([]models.Watchlist, error) {
	return s.watchlistRepo.GetUserWatchlist(userID)
}
