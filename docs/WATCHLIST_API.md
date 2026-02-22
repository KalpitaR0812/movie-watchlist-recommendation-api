# Watchlist API Documentation

## Overview

The Watchlist API provides endpoints for managing personal movie watchlists. Users can add movies to their watchlist, remove movies, and retrieve their complete watchlist. All endpoints require JWT authentication.

## Authentication

All watchlist endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

The token is validated by the authentication middleware, which extracts the user ID and injects it into the request context.

## API Endpoints

### Add Movie to Watchlist

**Endpoint**: `POST /api/v1/watchlist`

**Purpose**: Add a movie to the authenticated user's watchlist.

**Authentication**: Required (JWT Bearer Token)

**Request Body**:
```json
{
  "movie_id": "507f1f77bcf86cd799439011"
}
```

**Request Parameters**:
- `movie_id` (string, required): MongoDB ObjectID of the movie to add

**Response Examples**:

**Success (201 Created)**:
```json
{
  "message": "Movie added to watchlist successfully",
  "movie_id": "507f1f77bcf86cd799439011"
}
```

**Error Responses**:

**Bad Request (400)**:
```json
{
  "error": "Invalid movie ID format"
}
```

**Unauthorized (401)**:
```json
{
  "error": "User not authenticated"
}
```

**Conflict (409)**:
```json
{
  "error": "Movie is already in your watchlist"
}
```

**Not Found (404)**:
```json
{
  "error": "Movie not found"
}
```

**Implementation Details**:
- Validates movie ID format (MongoDB ObjectID)
- Verifies movie exists in the database
- Checks if movie is already in user's watchlist
- Creates watchlist entry with timestamp
- Returns success message with movie ID

### Remove Movie from Watchlist

**Endpoint**: `DELETE /api/v1/watchlist/{movieId}`

**Purpose**: Remove a movie from the authenticated user's watchlist.

**Authentication**: Required (JWT Bearer Token)

**Path Parameters**:
- `movieId` (string, required): MongoDB ObjectID of the movie to remove

**Response Examples**:

**Success (200 OK)**:
```json
{
  "message": "Movie removed from watchlist successfully",
  "movie_id": "507f1f77bcf86cd799439011"
}
```

**Error Responses**:

**Bad Request (400)**:
```json
{
  "error": "Invalid movie ID format"
}
```

**Unauthorized (401)**:
```json
{
  "error": "User not authenticated"
}
```

**Not Found (404)**:
```json
{
  "error": "Movie not found in your watchlist"
}
```

**Implementation Details**:
- Validates movie ID format from URL path
- Checks if movie exists in user's watchlist
- Removes watchlist entry
- Returns success message with removed movie ID

### Get User Watchlist

**Endpoint**: `GET /api/v1/watchlist`

**Purpose**: Retrieve all movies in the authenticated user's watchlist.

**Authentication**: Required (JWT Bearer Token)

**Query Parameters**: None

**Response Examples**:

**Success (200 OK)**:
```json
{
  "watchlist": [
    {
      "id": "507f1f77bcf86cd799439011",
      "added_at": "2023-12-01T10:30:00Z",
      "movie_id": "507f1f77bcf86cd799439012"
    },
    {
      "id": "507f1f77bcf86cd799439013",
      "added_at": "2023-12-02T14:15:00Z",
      "movie_id": "507f1f77bcf86cd799439014"
    }
  ],
  "count": 2
}
```

**Error Responses**:

**Unauthorized (401)**:
```json
{
  "error": "User not authenticated"
}
```

**Internal Server Error (500)**:
```json
{
  "error": "Failed to retrieve watchlist"
}
```

**Implementation Details**:
- Retrieves all watchlist entries for the authenticated user
- Returns watchlist items with basic information (ID, added date, movie ID)
- Includes total count of watchlist items
- Results are ordered by addition date (newest first)

### Get Watchlist with Movie Details

**Endpoint**: `GET /api/v1/watchlist/details`

**Purpose**: Retrieve user's watchlist with complete movie details.

**Authentication**: Required (JWT Bearer Token)

**Query Parameters**: None

**Response Examples**:

**Success (200 OK)**:
```json
{
  "watchlist": [
    {
      "watchlist_id": "507f1f77bcf86cd799439011",
      "added_at": "2023-12-01T10:30:00Z",
      "movie": {
        "_id": "507f1f77bcf86cd799439012",
        "imdb_id": "tt1375666",
        "title": "Inception",
        "year": "2010",
        "genre": "Action, Sci-Fi",
        "director": "Christopher Nolan",
        "plot": "A thief who steals corporate secrets...",
        "poster": "https://example.com/poster.jpg",
        "runtime": "148 min",
        "imdb_rating": "8.8",
        "cached_at": "2023-12-01T10:30:00Z",
        "created_at": "2023-12-01T10:30:00Z",
        "updated_at": "2023-12-01T10:30:00Z"
      }
    }
  ],
  "count": 1
}
```

**Error Responses**:

**Unauthorized (401)**:
```json
{
  "error": "User not authenticated"
}
```

**Internal Server Error (500)**:
```json
{
  "error": "Failed to retrieve watchlist"
}
```

**Implementation Details**:
- Retrieves watchlist entries and associated movie details
- Performs database join operation (watchlist + movies)
- Includes complete movie information for each watchlist item
- Skips items where movie details are not available
- Returns enriched data suitable for frontend display

## Data Model

### Watchlist Entry Structure

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

### Request/Response Models

**AddToWatchlistRequest**:
```go
type AddToWatchlistRequest struct {
    MovieID string `json:"movie_id" binding:"required"`
}
```

**WatchlistResponse**:
```go
type WatchlistResponse struct {
    ID       primitive.ObjectID `json:"id"`
    AddedAt  time.Time         `json:"added_at"`
    MovieID  primitive.ObjectID `json:"movie_id"`
}
```

**WatchlistWithDetailsResponse**:
```go
type WatchlistWithDetailsResponse struct {
    WatchlistID primitive.ObjectID `json:"watchlist_id"`
    AddedAt     time.Time         `json:"added_at"`
    Movie       *models.Movie     `json:"movie"`
}
```

## Business Logic

### Duplicate Prevention
- Each user can only add a specific movie to their watchlist once
- Duplicate attempts return 409 Conflict status
- Enforced through unique composite index on (user_id, movie_id)

### Movie Validation
- Movie must exist in the database before adding to watchlist
- Invalid movie IDs return 404 Not Found status
- Movie validation prevents orphaned watchlist entries

### Timestamp Management
- `added_at` field records when movie was added to watchlist
- `created_at` and `updated_at` fields track record lifecycle
- Timestamps use UTC timezone for consistency

### User Isolation
- Users can only access their own watchlist
- User ID extracted from JWT token for access control
- No cross-user data exposure possible

## Error Handling

### Validation Errors
- Invalid movie ID format: 400 Bad Request
- Missing required fields: 400 Bad Request
- Invalid JSON structure: 400 Bad Request

### Authentication Errors
- Missing token: 401 Unauthorized
- Invalid token: 401 Unauthorized
- Expired token: 401 Unauthorized

### Authorization Errors
- User not found: 401 Unauthorized
- Invalid user context: 401 Unauthorized

### Business Logic Errors
- Movie not found: 404 Not Found
- Movie already in watchlist: 409 Conflict
- Watchlist item not found: 404 Not Found

### System Errors
- Database connection issues: 500 Internal Server Error
- Repository query failures: 500 Internal Server Error
- Unexpected system errors: 500 Internal Server Error

## Performance Considerations

### Database Queries
- Watchlist queries use user_id index for efficient retrieval
- Movie details queries use movie_id index
- Composite indexes prevent duplicate entries

### Response Size
- Basic watchlist endpoint returns minimal data for fast responses
- Detailed watchlist endpoint includes full movie information
- Pagination can be implemented for large watchlists

### Caching Strategy
- Movie details are cached from OMDb API
- Watchlist data is real-time (no caching)
- Database connection pooling handles concurrent requests

## Security Considerations

### Access Control
- JWT tokens validated on every request
- User context injected prevents cross-user access
- Movie ID validation prevents injection attacks

### Data Validation
- All input validated using binding tags
- MongoDB ObjectID format validation
- SQL injection prevention through parameterized queries

### Rate Limiting
- No explicit rate limiting (future enhancement)
- Relies on infrastructure-level protection
- Can be implemented at middleware level

## Integration Points

### Authentication Middleware
- Extracts user ID from JWT token
- Validates token signature and expiration
- Injects user context into request

### Movie Service
- Validates movie existence
- Provides movie details for enriched responses
- Handles movie data caching

### Repository Layer
- Manages watchlist data persistence
- Enforces unique constraints
- Handles database operations

## Usage Examples

### Adding Multiple Movies
```bash
# Add first movie
curl -X POST http://localhost:8080/api/v1/watchlist \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"movie_id": "507f1f77bcf86cd799439011"}'

# Add second movie
curl -X POST http://localhost:8080/api/v1/watchlist \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"movie_id": "507f1f77bcf86cd799439012"}'
```

### Retrieving Watchlist
```bash
# Get basic watchlist
curl -X GET http://localhost:8080/api/v1/watchlist \
  -H "Authorization: Bearer <token>"

# Get watchlist with movie details
curl -X GET http://localhost:8080/api/v1/watchlist/details \
  -H "Authorization: Bearer <token>"
```

### Removing Movies
```bash
# Remove specific movie
curl -X DELETE http://localhost:8080/api/v1/watchlist/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <token>"
```

## Future Enhancements

### Pagination
- Implement pagination for large watchlists
- Add limit and offset parameters
- Include total count in responses

### Sorting Options
- Sort by addition date (newest/oldest)
- Sort by movie title (alphabetical)
- Sort by movie rating

### Filtering
- Filter by genre
- Filter by year
- Filter by rating range

### Batch Operations
- Add multiple movies in single request
- Remove multiple movies in single request
- Bulk operations for efficiency
