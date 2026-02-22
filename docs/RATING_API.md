# Rating API Documentation

## Overview

The Rating API provides endpoints for users to rate movies on a 1-5 star scale, update existing ratings, and retrieve rating history. The system enforces a one-rating-per-user-per-movie policy to maintain data integrity and support the recommendation engine.

## Authentication

All rating endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

The authentication middleware validates the token and extracts the user ID for rating operations.

## API Endpoints

### Rate a Movie

**Endpoint**: `POST /api/v1/ratings`

**Purpose**: Submit a new rating for a movie (1-5 stars).

**Authentication**: Required (JWT Bearer Token)

**Request Body**:
```json
{
  "movie_id": "507f1f77bcf86cd799439011",
  "rating": 5
}
```

**Request Parameters**:
- `movie_id` (string, required): MongoDB ObjectID of the movie to rate
- `rating` (integer, required): Rating value from 1 to 5 inclusive

**Validation Rules**:
- `movie_id`: Must be a valid MongoDB ObjectID format
- `rating`: Must be between 1 and 5 inclusive

**Response Examples**:

**Success (201 Created)**:
```json
{
  "message": "Movie rated successfully",
  "movie_id": "507f1f77bcf86cd799439011",
  "rating": 5,
  "stars": "★★★★★"
}
```

**Error Responses**:

**Bad Request (400)**:
```json
{
  "error": "Invalid movie ID format"
}
```

```json
{
  "error": "Rating must be between 1 and 5 stars"
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
  "error": "You have already rated this movie. Use the update endpoint to change your rating."
}
```

**Not Found (404)**:
```json
{
  "error": "Movie not found"
}
```

**Implementation Details**:
- Validates movie exists in database
- Checks if user has already rated the movie
- Creates new rating record with timestamp
- Returns star display representation
- Prevents duplicate ratings through unique constraint

### Update Movie Rating

**Endpoint**: `PUT /api/v1/ratings/{movieId}`

**Purpose**: Update an existing movie rating.

**Authentication**: Required (JWT Bearer Token)

**Path Parameters**:
- `movieId` (string, required): MongoDB ObjectID of the movie to update rating for

**Request Body**:
```json
{
  "rating": 4
}
```

**Request Parameters**:
- `rating` (integer, required): New rating value from 1 to 5 inclusive

**Validation Rules**:
- `rating`: Must be between 1 and 5 inclusive

**Response Examples**:

**Success (200 OK)**:
```json
{
  "message": "Rating updated successfully",
  "movie_id": "507f1f77bcf86cd799439011",
  "rating": 4,
  "stars": "★★★★☆"
}
```

**Error Responses**:

**Bad Request (400)**:
```json
{
  "error": "Invalid movie ID format"
}
```

```json
{
  "error": "Rating must be between 1 and 5 stars"
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
  "error": "You haven't rated this movie yet. Use the rate endpoint to add a rating."
}
```

**Implementation Details**:
- Validates movie ID format from URL path
- Checks if user has existing rating for the movie
- Updates rating value and timestamp
- Returns updated rating with star display
- Maintains rating history through timestamps

### Get User Ratings

**Endpoint**: `GET /api/v1/ratings`

**Purpose**: Retrieve all ratings submitted by the authenticated user.

**Authentication**: Required (JWT Bearer Token)

**Query Parameters**: None

**Response Examples**:

**Success (200 OK)**:
```json
{
  "ratings": [
    {
      "id": "507f1f77bcf86cd799439011",
      "movie_id": "507f1f77bcf86cd799439012",
      "rating": 5,
      "stars": "★★★★★",
      "created_at": "2023-12-01T10:30:00Z",
      "updated_at": "2023-12-01T10:30:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439013",
      "movie_id": "507f1f77bcf86cd799439014",
      "rating": 3,
      "stars": "★★★☆☆",
      "created_at": "2023-12-02T14:15:00Z",
      "updated_at": "2023-12-02T14:15:00Z"
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
  "error": "Failed to retrieve ratings"
}
```

**Implementation Details**:
- Retrieves all rating records for authenticated user
- Returns ratings with star display representation
- Includes creation and update timestamps
- Ordered by most recently updated first

### Get User Ratings with Movie Details

**Endpoint**: `GET /api/v1/ratings/details`

**Purpose**: Retrieve user ratings with complete movie information.

**Authentication**: Required (JWT Bearer Token)

**Query Parameters**: None

**Response Examples**:

**Success (200 OK)**:
```json
{
  "ratings": [
    {
      "rating_id": "507f1f77bcf86cd799439011",
      "rating": 5,
      "stars": "★★★★★",
      "created_at": "2023-12-01T10:30:00Z",
      "updated_at": "2023-12-01T10:30:00Z",
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
  "error": "Failed to retrieve ratings"
}
```

**Implementation Details**:
- Retrieves ratings and associated movie details
- Performs database join operation (ratings + movies)
- Includes complete movie information for each rating
- Skips ratings where movie details are unavailable
- Returns enriched data suitable for frontend display

### Get Specific Movie Rating

**Endpoint**: `GET /api/v1/ratings/{movieId}`

**Purpose**: Retrieve the authenticated user's rating for a specific movie.

**Authentication**: Required (JWT Bearer Token)

**Path Parameters**:
- `movieId` (string, required): MongoDB ObjectID of the movie

**Response Examples**:

**Success (200 OK)**:
```json
{
  "rating_id": "507f1f77bcf86cd799439011",
  "movie_id": "507f1f77bcf86cd799439012",
  "rating": 4,
  "stars": "★★★★☆",
  "created_at": "2023-12-01T10:30:00Z",
  "updated_at": "2023-12-01T10:30:00Z"
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
  "error": "Rating not found"
}
```

**Implementation Details**:
- Validates movie ID format from URL path
- Retrieves specific rating for user and movie combination
- Returns rating with star display and timestamps
- Efficient lookup using composite index

### Delete Movie Rating

**Endpoint**: `DELETE /api/v1/ratings/{movieId}`

**Purpose**: Delete a user's rating for a specific movie.

**Authentication**: Required (JWT Bearer Token)

**Path Parameters**:
- `movieId` (string, required): MongoDB ObjectID of the movie

**Response Examples**:

**Success (200 OK)**:
```json
{
  "message": "Rating deleted successfully",
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
  "error": "Rating not found"
}
```

**Implementation Details**:
- Validates movie ID format from URL path
- Checks if rating exists for user and movie combination
- Removes rating record from database
- Returns confirmation message

## Data Model

### Rating Structure

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

### Request/Response Models

**RateMovieRequest**:
```go
type RateMovieRequest struct {
    MovieID string `json:"movie_id" binding:"required"`
    Rating  int    `json:"rating" binding:"required,min=1,max=5"`
}
```

**UpdateRatingRequest**:
```go
type UpdateRatingRequest struct {
    Rating int `json:"rating" binding:"required,min=1,max=5"`
}
```

**RatingResponse**:
```go
type RatingResponse struct {
    ID        primitive.ObjectID `json:"id"`
    MovieID   primitive.ObjectID `json:"movie_id"`
    Rating    int               `json:"rating"`
    Stars     string            `json:"stars"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
}
```

## Duplicate Rating Prevention

### Business Rule
Each user can rate a specific movie only once. This policy is enforced at multiple levels:

### Database Level
- **Unique Composite Index**: `{ "user_id": 1, "movie_id": 1 }`
- Prevents duplicate rating records in the database
- Database returns error on duplicate insertion attempts

### Application Level
- **Pre-insertion Check**: Service layer checks for existing rating before creation
- **User-friendly Error**: Returns 409 Conflict with descriptive message
- **Update Guidance**: Suggests using update endpoint for existing ratings

### Implementation Logic
```go
// Check for existing rating
existing, err := s.ratingRepo.GetUserRating(userID, movieID)
if err == nil && existing != nil {
    return errors.New("user has already rated this movie")
}

// Create new rating if none exists
return s.ratingRepo.Create(newRating)
```

## Rating System Impact on Recommendations

### Recommendation Engine Integration
The rating system is fundamental to the recommendation engine:

### Genre Preference Analysis
- **High-rated Genres**: Movies rated 4+ stars identify user preferences
- **Genre Weighting**: Higher ratings increase genre preference weight
- **Recommendation Scoring**: Preferred genres receive higher recommendation scores

### Exclusion Logic
- **Already Rated**: Rated movies are excluded from recommendations
- **Watchlist Integration**: Watchlist movies also excluded
- **User History**: Complete user history considered for personalization

### Recommendation Algorithm
1. **Analyze User Ratings**: Extract genres from 4+ star ratings
2. **Calculate Genre Scores**: Weight genres by rating frequency and value
3. **Find Matching Movies**: Search movies in preferred genres
4. **Exclude User History**: Remove already rated/watchlisted movies
5. **Score and Rank**: Calculate recommendation scores based on genre matching

### Rating Quality Considerations
- **Minimum Ratings**: Users need sufficient ratings for accurate recommendations
- **Rating Distribution**: Variety of ratings improves recommendation accuracy
- **Temporal Factors**: Recent ratings may have higher weight

## Star Display System

### Visual Representation
The system provides both numeric and visual star representations:

### Star Conversion Logic
```go
func getStarDisplay(rating int) string {
    stars := ""
    for i := 1; i <= 5; i++ {
        if i <= rating {
            stars += "★"  // Filled star
        } else {
            stars += "☆"  // Empty star
        }
    }
    return stars
}
```

### Rating to Star Mapping
- **1 star**: "☆☆☆☆☆"
- **2 stars**: "★★☆☆☆"
- **3 stars**: "★★★☆☆"
- **4 stars**: "★★★★☆"
- **5 stars**: "★★★★★"

### Unicode Considerations
- Uses Unicode star characters for consistent display
- Compatible with modern web browsers and mobile devices
- Provides immediate visual feedback for rating levels

## Error Handling

### Validation Errors
- **Invalid Movie ID**: 400 Bad Request with format validation message
- **Invalid Rating Value**: 400 Bad Request with range validation message
- **Missing Required Fields**: 400 Bad Request with field specification

### Authentication Errors
- **Missing Token**: 401 Unauthorized
- **Invalid Token**: 401 Unauthorized
- **Expired Token**: 401 Unauthorized

### Business Logic Errors
- **Movie Not Found**: 404 Not Found
- **Rating Already Exists**: 409 Conflict with update suggestion
- **Rating Not Found**: 404 Not Found (for update/delete operations)

### System Errors
- **Database Connection**: 500 Internal Server Error
- **Repository Failures**: 500 Internal Server Error
- **Unexpected Errors**: 500 Internal Server Error

## Performance Considerations

### Database Optimization
- **Composite Index**: Fast lookup of user-movie rating combinations
- **User Index**: Efficient retrieval of user's rating history
- **Movie Index**: Quick access to all ratings for specific movies

### Response Optimization
- **Pagination**: Can be implemented for large rating histories
- **Field Selection**: Basic endpoints return minimal data
- **Caching**: Movie details cached from OMDb API

### Query Efficiency
- **Single Record Lookups**: Efficient for specific movie ratings
- **Batch Operations**: User history retrieval optimized with indexes
- **Join Operations**: Movie details joined only when requested

## Security Considerations

### Access Control
- **JWT Validation**: Token validated on every request
- **User Isolation**: Users can only access their own ratings
- **Movie Validation**: Prevents rating of non-existent movies

### Data Integrity
- **Input Validation**: All inputs validated using binding tags
- **Type Safety**: Strong typing prevents injection attacks
- **Constraint Enforcement**: Database constraints prevent data corruption

### Rate Limiting
- **No Explicit Limits**: Relies on infrastructure protection
- **Future Enhancement**: Can implement per-user rating limits
- **Abuse Prevention**: Duplicate prevention reduces spam ratings

## Integration Points

### Authentication Middleware
- **Token Extraction**: JWT token extracted from Authorization header
- **User Context**: User ID injected into request context
- **Access Control**: Middleware enforces authentication requirements

### Movie Service
- **Movie Validation**: Verifies movie existence before rating
- **Details Retrieval**: Provides movie information for enriched responses
- **Cache Integration**: Uses cached movie data for efficiency

### Recommendation Service
- **Rating Analysis**: Consumes rating data for preference analysis
- **Genre Extraction**: Processes ratings to identify user preferences
- **Exclusion Logic**: Uses ratings to filter recommendation results

## Usage Examples

### Rating Workflow
```bash
# Rate a movie
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"movie_id": "507f1f77bcf86cd799439011", "rating": 5}'

# Update the rating
curl -X PUT http://localhost:8080/api/v1/ratings/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"rating": 4}'

# Get rating history
curl -X GET http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer <token>"

# Get rating with movie details
curl -X GET http://localhost:8080/api/v1/ratings/details \
  -H "Authorization: Bearer <token>"
```

### Error Handling Examples
```bash
# Attempt duplicate rating
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"movie_id": "507f1f77bcf86cd799439011", "rating": 3}'
# Response: 409 Conflict - "You have already rated this movie"

# Invalid rating value
curl -X POST http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"movie_id": "507f1f77bcf86cd799439011", "rating": 6}'
# Response: 400 Bad Request - "Rating must be between 1 and 5 stars"
```

## Future Enhancements

### Advanced Rating Features
- **Half-star Ratings**: Support for 0.5 star increments
- **Rating Tags**: Add tags or notes to ratings
- **Rating History**: Track rating changes over time
- **Rating Analytics**: Provide user rating statistics

### Recommendation Enhancements
- **Collaborative Filtering**: Use similar users' ratings
- **Content-based Filtering**: Analyze movie attributes
- **Hybrid Approach**: Combine multiple recommendation strategies
- **Real-time Updates**: Update recommendations as users rate movies

### Performance Improvements
- **Rating Caching**: Cache frequently accessed rating data
- **Batch Processing**: Process multiple ratings efficiently
- **Analytics Precomputation**: Precompute recommendation data
- **Database Optimization**: Advanced indexing strategies
