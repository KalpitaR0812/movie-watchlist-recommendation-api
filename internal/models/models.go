package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string            `bson:"username" json:"username"`
	Email     string            `bson:"email" json:"email"`
	Password  string            `bson:"password" json:"-"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	IMDbID      string            `bson:"imdb_id" json:"imdb_id"`
	Title       string            `bson:"title" json:"title"`
	Year        string            `bson:"year" json:"year"`
	Genre       string            `bson:"genre" json:"genre"`
	Director    string            `bson:"director" json:"director"`
	Plot        string            `bson:"plot" json:"plot"`
	Poster      string            `bson:"poster" json:"poster"`
	Runtime     string            `bson:"runtime" json:"runtime"`
	IMDbRating  string            `bson:"imdb_rating" json:"imdb_rating"`
	CachedAt    time.Time         `bson:"cached_at" json:"cached_at"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type Watchlist struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	AddedAt   time.Time         `bson:"added_at" json:"added_at"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type Rating struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	Rating    int               `bson:"rating" json:"rating"` // Changed to int for 1-5 star system
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}
