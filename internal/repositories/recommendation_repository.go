package repositories

import (
	"context"
	"movie-watchlist/internal/database"
	"movie-watchlist/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecommendationRepository struct {
	db *database.MongoDB
}

func NewRecommendationRepository(db *database.MongoDB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// GetHighRatedGenres fetches genres from ratings where rating >= 4
func (r *RecommendationRepository) GetHighRatedGenres(userID primitive.ObjectID, threshold int) ([]string, error) {
	ctx := context.Background()
	ratingsCollection := r.db.GetCollection("ratings")
	
	// Aggregation pipeline to find genres rated >= threshold
	pipeline := []bson.M{
		// Stage 1: Match ratings by user and rating threshold
		{
			"$match": bson.M{
				"user_id": userID,
				"rating":  bson.M{"$gte": threshold},
			},
		},
		// Stage 2: Lookup movie details to get genre
		{
			"$lookup": bson.M{
				"from":         "movies",
				"localField":   "movie_id",
				"foreignField": "_id",
				"as":           "movie",
			},
		},
		// Stage 3: Unwind the movie array
		{
			"$unwind": "$movie",
		},
		// Stage 4: Split genre string into array (handle multiple genres)
		{
			"$project": bson.M{
				"genres": bson.M{
					"$split": bson.A{"$movie.genre", ","},
				},
			},
		},
		// Stage 5: Unwind genres array
		{
			"$unwind": "$genres",
		},
		// Stage 6: Trim whitespace from genre names
		{
			"$project": bson.M{
				"genre": bson.M{
					"$trim": bson.M{"input": "$genres"},
				},
			},
		},
		// Stage 7: Filter out empty genres
		{
			"$match": bson.M{
				"genre": bson.M{"$ne": ""},
			},
		},
		// Stage 8: Group by genre and count occurrences
		{
			"$group": bson.M{
				"_id":   "$genre",
				"count": bson.M{"$sum": 1},
			},
		},
		// Stage 9: Sort by count (most frequent first)
		{
			"$sort": bson.M{"count": -1},
		},
		// Stage 10: Extract genre names
		{
			"$project": bson.M{
				"_id":   0,
				"genre": "$_id",
			},
		},
	}
	
	cursor, err := ratingsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		Genre string `bson:"genre"`
		Count int    `bson:"count"`
	}
	
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	// Extract unique genres in order of preference
	genres := make([]string, 0, len(results))
	for _, result := range results {
		genres = append(genres, result.Genre)
	}
	
	return genres, nil
}

// GetRatedMovieIDs fetches movie IDs from ratings collection
func (r *RecommendationRepository) GetRatedMovieIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("ratings")
	
	// Simple find query to get all movie IDs for a user
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

// GetWatchlistMovieIDs fetches movie IDs from watchlist collection
func (r *RecommendationRepository) GetWatchlistMovieIDs(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("watchlists")
	
	// Simple find query to get all movie IDs from user's watchlist
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var watchlists []models.Watchlist
	if err := cursor.All(ctx, &watchlists); err != nil {
		return nil, err
	}
	
	// Extract movie IDs
	movieIDs := make([]primitive.ObjectID, len(watchlists))
	for i, watchlist := range watchlists {
		movieIDs[i] = watchlist.MovieID
	}
	
	return movieIDs, nil
}

// GetMoviesToExclude combines rated and watchlist movie IDs
func (r *RecommendationRepository) GetMoviesToExclude(userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	// Get rated movie IDs
	ratedIDs, err := r.GetRatedMovieIDs(userID)
	if err != nil {
		return nil, err
	}
	
	// Get watchlist movie IDs
	watchlistIDs, err := r.GetWatchlistMovieIDs(userID)
	if err != nil {
		return nil, err
	}
	
	// Combine and deduplicate
	excludeMap := make(map[primitive.ObjectID]bool)
	for _, id := range ratedIDs {
		excludeMap[id] = true
	}
	for _, id := range watchlistIDs {
		excludeMap[id] = true
	}
	
	// Convert back to slice
	excludeIDs := make([]primitive.ObjectID, 0, len(excludeMap))
	for id := range excludeMap {
		excludeIDs = append(excludeIDs, id)
	}
	
	return excludeIDs, nil
}

// GetMoviesByGenreExcludingIDs fetches movies by genre excluding specified ObjectIDs
func (r *RecommendationRepository) GetMoviesByGenreExcludingIDs(genre string, excludeIDs []primitive.ObjectID, limit int) ([]models.Movie, error) {
	ctx := context.Background()
	collection := r.db.GetCollection("movies")
	
	// Build query filter
	filter := bson.M{
		"genre": bson.M{"$regex": genre, "$options": "i"}, // Case-insensitive genre match
	}
	
	// Add exclusion filter if there are IDs to exclude
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}
	
	// Find movies with limit
	findOptions := options.Find()
	if limit > 0 {
		findOptions.SetLimit(int64(limit))
	}
	findOptions.SetSort(bson.D{{Key: "imdb_rating", Value: -1}}) // Sort by IMDb rating descending
	
	cursor, err := collection.Find(ctx, filter, findOptions)
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

// GetRecommendationMovies is a comprehensive method that gets movies for recommendations
func (r *RecommendationRepository) GetRecommendationMovies(userID primitive.ObjectID, genres []string, limit int) ([]models.Movie, error) {
	ctx := context.Background()
	moviesCollection := r.db.GetCollection("movies")
	
	// Get movies to exclude (rated + watchlist)
	excludeIDs, err := r.GetMoviesToExclude(userID)
	if err != nil {
		return nil, err
	}
	
	// Build aggregation pipeline for genre-based recommendations
	pipeline := []bson.M{
		// Stage 1: Match movies that are not in exclude list and have specified genres
		{
			"$match": bson.M{
				"_id": bson.M{"$nin": excludeIDs},
				"$or": buildGenreMatchPipeline(genres),
			},
		},
		// Stage 2: Sort by IMDb rating (highest first)
		{
			"$sort": bson.M{"imdb_rating": -1},
		},
		// Stage 3: Limit results
		{
			"$limit": limit,
		},
	}
	
	cursor, err := moviesCollection.Aggregate(ctx, pipeline)
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

// buildGenreMatchPipeline creates $or conditions for genre matching
func buildGenreMatchPipeline(genres []string) []bson.M {
	if len(genres) == 0 {
		return []bson.M{}
	}
	
	genreConditions := make([]bson.M, len(genres))
	for i, genre := range genres {
		genreConditions[i] = bson.M{"genre": bson.M{"$regex": genre, "$options": "i"}}
	}
	
	return genreConditions
}

// GetMovieCountByGenre returns count of movies per genre (excluding user's movies)
func (r *RecommendationRepository) GetMovieCountByGenre(userID primitive.ObjectID, genres []string) (map[string]int64, error) {
	ctx := context.Background()
	moviesCollection := r.db.GetCollection("movies")
	
	// Get movies to exclude
	excludeIDs, err := r.GetMoviesToExclude(userID)
	if err != nil {
		return nil, err
	}
	
	// Build aggregation pipeline to count movies by genre
	pipeline := []bson.M{
		// Stage 1: Match movies not in exclude list
		{
			"$match": bson.M{
				"_id": bson.M{"$nin": excludeIDs},
				"$or": buildGenreMatchPipeline(genres),
			},
		},
		// Stage 2: Group by genre and count
		{
			"$group": bson.M{
				"_id":   "$genre",
				"count": bson.M{"$sum": 1},
			},
		},
		// Stage 3: Format results
		{
			"$project": bson.M{
				"genre": "$_id",
				"count": "$count",
			},
		},
	}
	
	cursor, err := moviesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var results []struct {
		Genre string `bson:"genre"`
		Count int64  `bson:"count"`
	}
	
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	
	// Convert to map
	genreCounts := make(map[string]int64)
	for _, result := range results {
		genreCounts[result.Genre] = result.Count
	}
	
	return genreCounts, nil
}
