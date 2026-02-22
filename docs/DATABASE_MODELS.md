# Database Models Documentation

## Overview

This document describes the data models used in the Movie Watchlist & Recommendation System. The system uses MongoDB as the primary database with MongoDB ObjectID as the primary key for all collections.

## Data Models

### User Model

**Collection**: `users`

**Purpose**: Stores user authentication and profile information.

```go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username  string            `bson:"username" json:"username"`
    Email     string            `bson:"email" json:"email"`
    Password  string            `bson:"password" json:"-"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}
```

**Field Descriptions**:
- `ID`: MongoDB ObjectID serving as primary key
- `Username`: Unique username for user identification
- `Email`: Unique email address for authentication
- `Password`: Hashed password (bcrypt), excluded from JSON responses
- `CreatedAt`: Timestamp when user account was created
- `UpdatedAt`: Timestamp when user account was last modified

**Constraints**:
- Email must be unique across all users
- Username must be unique across all users
- Password is stored using bcrypt hashing algorithm

### Movie Model

**Collection**: `movies`

**Purpose**: Stores cached movie data fetched from OMDb API.

```go
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
```

**Field Descriptions**:
- `ID`: MongoDB ObjectID serving as primary key
- `IMDbID`: Unique identifier from OMDb database
- `Title`: Movie title
- `Year`: Release year
- `Genre`: Comma-separated genre string (e.g., "Action, Sci-Fi")
- `Director`: Director name(s)
- `Plot`: Movie plot summary
- `Poster`: URL to movie poster image
- `Runtime`: Movie duration
- `IMDbRating`: IMDb rating (as string)
- `CachedAt`: Timestamp when movie data was cached from OMDb
- `CreatedAt`: Timestamp when record was created
- `UpdatedAt`: Timestamp when record was last modified

**Constraints**:
- IMDbID must be unique across all movies
- Genre field is required for recommendation functionality
- Movie data is cached to reduce external API calls

### Watchlist Model

**Collection**: `watchlists`

**Purpose**: Tracks movies added to users' personal watchlists.

```go
type Watchlist struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
    MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
    AddedAt   time.Time         `bson:"added_at" json:"added_at"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}
```

**Field Descriptions**:
- `ID`: MongoDB ObjectID serving as primary key
- `UserID`: Foreign key reference to User collection
- `MovieID`: Foreign key reference to Movie collection
- `AddedAt`: Timestamp when movie was added to watchlist
- `CreatedAt`: Timestamp when record was created
- `UpdatedAt`: Timestamp when record was last modified

**Constraints**:
- Combination of UserID and MovieID must be unique
- A user cannot add the same movie to watchlist multiple times

### Rating Model

**Collection**: `ratings`

**Purpose**: Stores user ratings for movies (1-5 star system).

```go
type Rating struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
    MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
    Rating    int               `bson:"rating" json:"rating"`
    CreatedAt time.Time         `bson:"created_at" json:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}
```

**Field Descriptions**:
- `ID`: MongoDB ObjectID serving as primary key
- `UserID`: Foreign key reference to User collection
- `MovieID`: Foreign key reference to Movie collection
- `Rating`: Integer rating from 1 to 5 stars
- `CreatedAt`: Timestamp when rating was created
- `UpdatedAt`: Timestamp when rating was last modified

**Constraints**:
- Combination of UserID and MovieID must be unique
- Rating value must be between 1 and 5 inclusive
- Users can only rate a movie once (must update existing rating)

## Relationships Between Collections

### User-Related Relationships
- **User → Watchlist**: One-to-many relationship (one user can have many watchlist items)
- **User → Rating**: One-to-many relationship (one user can rate many movies)

### Movie-Related Relationships
- **Movie → Watchlist**: One-to-many relationship (one movie can be in many users' watchlists)
- **Movie → Rating**: One-to-many relationship (one movie can be rated by many users)

### Cross-Collection Relationships
- **Watchlist items reference both User and Movie collections**
- **Rating records reference both User and Movie collections**

## Indexing Strategy

### User Collection Indexes
- **Email Index**: `{ "email": 1 }` - Unique index for fast user lookup by email
- **Username Index**: `{ "username": 1 }` - Unique index for fast user lookup by username

### Movie Collection Indexes
- **IMDbID Index**: `{ "imdb_id": 1 }` - Unique index for fast movie lookup by IMDb ID
- **Title Text Index**: `{ "title": "text" }` - Text index for movie search functionality
- **Genre Index**: `{ "genre": 1 }` - Index for genre-based recommendations

### Watchlist Collection Indexes
- **User-Movie Composite Index**: `{ "user_id": 1, "movie_id": 1 }` - Unique index preventing duplicates
- **User Index**: `{ "user_id": 1 }` - Index for fetching user's watchlist

### Rating Collection Indexes
- **User-Movie Composite Index**: `{ "user_id": 1, "movie_id": 1 }` - Unique index preventing duplicate ratings
- **User Index**: `{ "user_id": 1 }` - Index for fetching user's ratings
- **Movie Index**: `{ "movie_id": 1 }` - Index for fetching movie ratings
- **Rating Index**: `{ "rating": 1 }` - Index for recommendation calculations

## Data Integrity Considerations

### Referential Integrity
- The system maintains logical referential integrity through application-level validation
- Foreign key references are validated at the application layer
- Orphaned records are prevented through proper error handling

### Data Consistency
- Timestamps are automatically managed for audit trails
- Updates modify the `UpdatedAt` field to track record changes
- Caching timestamps (`CachedAt`) track when external data was fetched

### Performance Considerations
- Indexes are strategically placed to support common query patterns
- Composite indexes prevent duplicate entries while enabling fast lookups
- Text indexes support efficient movie search functionality

## Schema Evolution

The current schema supports:
- User authentication and management
- Movie data caching from external APIs
- Personal watchlist functionality
- Rating system with duplicate prevention
- Recommendation engine requirements

Future extensions might include:
- User preferences and settings
- Social features (following other users)
- Advanced filtering and sorting options
- Review and comment systems
