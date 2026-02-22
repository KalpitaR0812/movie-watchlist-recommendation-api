package database

import (
	"context"
	"fmt"
	"log"
	
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(mongoURI string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database name from URI or use default
	dbName := "movie_watchlist"
	if mongoURI[len(mongoURI)-1] == '/' {
		dbName = "movie_watchlist"
	}

	database := &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}

	// Create indexes
	if err := database.createIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	return database, nil
}

func (db *MongoDB) createIndexes(ctx context.Context) error {
	// Users collection indexes
	usersCollection := db.Database.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Movies collection indexes
	moviesCollection := db.Database.Collection("movies")
	_, err = moviesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "imdb_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "title", Value: 1}}},
		{Keys: bson.D{{Key: "genre", Value: 1}}},
		{Keys: bson.D{{Key: "cached_at", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create movies indexes: %w", err)
	}

	// Watchlists collection indexes
	watchlistsCollection := db.Database.Collection("watchlists")
	_, err = watchlistsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "movie_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "movie_id", Value: 1}}},
		{Keys: bson.D{{Key: "added_at", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create watchlists indexes: %w", err)
	}

	// Ratings collection indexes
	ratingsCollection := db.Database.Collection("ratings")
	_, err = ratingsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "movie_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "movie_id", Value: 1}}},
		{Keys: bson.D{{Key: "rating", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("failed to create ratings indexes: %w", err)
	}

	return nil
}

func (db *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.Client.Disconnect(ctx)
}

// Helper function to get collection
func (db *MongoDB) GetCollection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}
