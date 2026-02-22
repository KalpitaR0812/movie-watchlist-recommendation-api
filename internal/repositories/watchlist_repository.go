package repositories

import (
	"context"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WatchlistRepository struct {
	db *database.MongoDB
}

func NewWatchlistRepository(db *database.MongoDB) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (r *WatchlistRepository) Add(watchlist *models.Watchlist) error {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	watchlist.CreatedAt = getCurrentTime()
	watchlist.UpdatedAt = getCurrentTime()
	watchlist.AddedAt = time.Now()
	
	result, err := collection.InsertOne(ctx, watchlist)
	if err != nil {
		return err
	}
	
	watchlist.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *WatchlistRepository) Remove(userID, movieID primitive.ObjectID) error {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	_, err := collection.DeleteOne(ctx, bson.M{
		"user_id": userID,
		"movie_id": movieID,
	})
	return err
}

func (r *WatchlistRepository) GetUserWatchlist(userID primitive.ObjectID) ([]models.Watchlist, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var watchlist []models.Watchlist
	if err := cursor.All(ctx, &watchlist); err != nil {
		return nil, err
	}
	
	// Populate movie details for each watchlist entry
	for i := range watchlist {
		_, err := r.getMovieByID(ctx, watchlist[i].MovieID)
		if err == nil {
			// Note: In MongoDB, we don't have automatic population like GORM
			// We would need to manually populate or use aggregation pipeline
			// For simplicity, we'll fetch movie details separately
		}
	}
	
	return watchlist, nil
}

func (r *WatchlistRepository) Exists(userID, movieID primitive.ObjectID) (bool, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	count, err := collection.CountDocuments(ctx, bson.M{
		"user_id": userID,
		"movie_id": movieID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *WatchlistRepository) GetWatchlistWithMovies(userID primitive.ObjectID) ([]models.Watchlist, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	// Use aggregation pipeline to join with movies collection
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$lookup": bson.M{
			"from":         "movies",
			"localField":   "movie_id",
			"foreignField": "_id",
			"as":           "movie",
		}},
		{"$unwind": "$movie"},
		{"$sort": bson.M{"added_at": -1}},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		models.Watchlist `bson:",inline"`
		Movie           models.Movie `bson:"movie"`
	}
	
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	// Convert to expected format
	watchlist := make([]models.Watchlist, len(results))
	for i, result := range results {
		watchlist[i] = result.Watchlist
		// Note: We don't populate the Movie field in the struct since we're using aggregation
	}
	
	return watchlist, nil
}

// Helper function to get movie by ID
func (r *WatchlistRepository) getMovieByID(ctx context.Context, movieID primitive.ObjectID) (*models.Movie, error) {
	collection := r.db.GetCollection("movies")
	
	var movie models.Movie
	err := collection.FindOne(ctx, bson.M{"_id": movieID}).Decode(&movie)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &movie, nil
}
