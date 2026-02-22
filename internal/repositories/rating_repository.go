package repositories

import (
	"context"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/models"
	

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RatingRepository struct {
	db *database.MongoDB
}

func NewRatingRepository(db *database.MongoDB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Create(rating *models.Rating) error {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	rating.CreatedAt = getCurrentTime()
	rating.UpdatedAt = getCurrentTime()
	
	result, err := collection.InsertOne(ctx, rating)
	if err != nil {
		return err
	}
	
	rating.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *RatingRepository) Update(userID, movieID primitive.ObjectID, rating int) error {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	update := bson.M{
		"$set": bson.M{
			"rating":     rating,
			"updated_at": getCurrentTime(),
		},
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{
		"user_id":  userID,
		"movie_id": movieID,
	}, update)
	
	return err
}

func (r *RatingRepository) GetUserRating(userID, movieID primitive.ObjectID) (*models.Rating, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	var rating models.Rating
	err := collection.FindOne(ctx, bson.M{
		"user_id":  userID,
		"movie_id": movieID,
	}).Decode(&rating)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &rating, nil
}

func (r *RatingRepository) GetUserRatings(userID primitive.ObjectID) ([]models.Rating, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var ratings []models.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}
	
	return ratings, nil
}

func (r *RatingRepository) GetHighRatedGenres(userID primitive.ObjectID, threshold int) ([]string, error) {
	ctx := context.Background()
	ratingsCollection := r.db.GetCollection("ratings")
	
	// Use aggregation pipeline to join ratings and movies
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id": userID,
			"rating":  bson.M{"$gte": threshold},
		}},
		{"$lookup": bson.M{
			"from":         "movies",
			"localField":   "movie_id",
			"foreignField": "_id",
			"as":           "movie",
		}},
		{"$unwind": "$movie"},
		{"$project": bson.M{
			"genre": "$movie.genre",
		}},
		{"$group": bson.M{
			"_id":   "$genre",
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
	}
	
	cursor, err := ratingsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		Genre string `bson:"_id"`
		Count int    `bson:"count"`
	}
	
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	// Extract unique genres
	genres := make([]string, len(results))
	for i, result := range results {
		genres[i] = result.Genre
	}
	
	return genres, nil
}

func (r *RatingRepository) GetRatedMovieIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var ratings []models.Rating
	if err := cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}
	
	// Extract movie IDs
	movieIDs := make([]primitive.ObjectID, len(ratings))
	for i, rating := range ratings {
		movieIDs[i] = rating.MovieID
	}
	
	return movieIDs, nil
}
