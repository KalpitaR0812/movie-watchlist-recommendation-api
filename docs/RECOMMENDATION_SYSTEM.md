# Recommendation System Documentation

## Overview

The Movie Watchlist system implements a rule-based recommendation engine that provides personalized movie suggestions based on user rating history. The system uses a deterministic approach without machine learning, making it suitable for academic evaluation and transparent decision-making.

## Recommendation Philosophy

### Rule-Based Approach
The recommendation system follows a rule-based methodology rather than machine learning for several reasons:

- **Transparency**: Every recommendation can be explained with clear rules
- **Deterministic**: Same user data always produces same recommendations
- **Academic Suitability**: Logic can be easily analyzed and evaluated
- **Maintenance**: Rules can be adjusted without retraining models
- **Performance**: Fast execution without complex computations

### Core Principle
Users prefer movies similar to those they have rated highly (4+ stars). The system identifies preferred genres and suggests movies from those genres that the user hasn't seen yet.

## Recommendation Algorithm

### Step 1: User Preference Analysis

#### High-Rated Genre Identification
```go
// Get genres from movies rated 4+ stars
highRatedGenres, err := s.recommendationRepo.GetHighRatedGenres(userID, 4)
```

**Process**:
1. Query user's rating history
2. Filter ratings with value ≥ 4 stars
3. Extract genres from highly-rated movies
4. Count genre occurrences
5. Return sorted list of preferred genres

**Example Output**:
```
["Action", "Sci-Fi", "Thriller"]
```

#### Genre Scoring Logic
- **Frequency Weight**: More ratings in genre = higher preference
- **Rating Weight**: Higher average rating = stronger preference
- **Minimum Threshold**: Only genres with sufficient data considered

### Step 2: Exclusion Set Creation

#### Already Rated Movies
```go
ratedMovies, err := s.ratingRepo.GetUserRatings(userID)
```

**Process**:
1. Retrieve all user's movie ratings
2. Extract movie IDs from rating records
3. Add to exclusion set

#### Watchlist Movies
```go
watchlistMovies, err := s.watchlistRepo.GetUserWatchlist(userID)
```

**Process**:
1. Retrieve user's watchlist
2. Extract movie IDs from watchlist entries
3. Add to exclusion set

#### Combined Exclusion Logic
```go
excludeMovieIDs := append(ratedMovieIDs, watchlistMovieIDs...)
```

**Purpose**: Prevent recommending movies the user already knows about or has expressed interest in.

### Step 3: Candidate Movie Selection

#### Genre-Based Filtering
```go
recommendations, err := s.recommendationRepo.GetMoviesByGenres(
    highRatedGenres, 
    excludeMovieIDs, 
    limit
)
```

**MongoDB Aggregation Pipeline**:
```javascript
[
  // Match movies in preferred genres
  {
    "$match": {
      "genre": { "$regex": "Action|Sci-Fi", "$options": "i" }
    }
  },
  // Exclude already rated/watchlisted movies
  {
    "$match": {
      "_id": { "$nin": [ratedMovieIDs, watchlistMovieIDs] }
    }
  },
  // Limit results
  { "$limit": limit }
]
```

**Process**:
1. Find movies matching preferred genres
2. Exclude movies in exclusion set
3. Limit to requested number of recommendations
4. Return candidate movies

### Step 4: Scoring and Ranking

#### Recommendation Score Calculation
```go
func calculateRecommendationScore(movie *models.Movie, preferredGenres []string) float64 {
    score := 0.5 // Base score
    
    // Add points for genre matching (30% weight)
    movieGenres := parseGenres(movie.Genre)
    for _, preferred := range preferredGenres {
        for _, movieGenre := range movieGenres {
            if genresMatch(preferred, movieGenre) {
                score += 0.3
                break
            }
        }
    }
    
    // Add points for high IMDb rating (20% weight)
    if movie.IMDbRating >= "8.0" {
        score += 0.2
    } else if movie.IMDbRating >= "7.0" {
        score += 0.1
    }
    
    return min(score, 1.0)
}
```

**Scoring Components**:
- **Base Score**: 50% (all movies start here)
- **Genre Matching**: 30% (preferred genres get bonus)
- **IMDb Rating**: 20% (highly-rated movies get bonus)

#### Confidence Levels
```go
func getConfidenceLevel(score float64) string {
    if score >= 0.8 {
        return "high"
    } else if score >= 0.6 {
        return "medium"
    }
    return "low"
}
```

**Confidence Tiers**:
- **High (≥0.8)**: Strong genre match + high IMDb rating
- **Medium (≥0.6)**: Moderate match or good rating
- **Low (<0.6)**: Weak match or fallback recommendation

### Step 5: Fallback Strategy

#### Popular Movies Fallback
```go
if len(recommendations) < limit {
    remaining := limit - len(recommendations)
    popularMovies, err := getPopularMoviesExcluding(remaining, excludeMovieIDs, recommendations)
    recommendations = append(recommendations, popularMovies...)
}
```

**Fallback Logic**:
1. Triggered when insufficient genre-based recommendations
2. Uses popular movies (high IMDb ratings)
3. Still excludes user's rated/watchlisted movies
4. Ensures minimum number of recommendations

#### Popular Movie Selection
- **IMDb Rating Sort**: Highest rated movies first
- **Exclusion Applied**: Remove already known movies
- **Quantity Limited**: Fill remaining recommendation slots

## Deterministic Behavior

### Consistency Guarantees
The system ensures deterministic behavior through:

#### Fixed Algorithm
- Same input always produces same output
- No random elements in recommendation logic
- Predictable scoring and ranking

#### Stable Data Sources
- User ratings provide consistent preference data
- Cached movie data ensures stable movie attributes
- Exclusion sets prevent recommendation changes

#### Temporal Stability
- Recommendations change only when user ratings change
- No time-based randomization
- Consistent ordering within same score ranges

### Reproducibility
```go
// Same user, same time = same recommendations
recommendations1 := getRecommendations(userID, 10)
recommendations2 := getRecommendations(userID, 10)
// recommendations1 == recommendations2 (always true)
```

## Academic Evaluation Suitability

### Transparent Logic
Every recommendation can be explained:

```
Recommended "Inception" because:
- You rated "The Matrix" 5 stars (Sci-Fi genre)
- You rated "Dark Knight" 4 stars (Action genre)
- "Inception" matches your preferred genres (Action, Sci-Fi)
- High IMDb rating (8.8) increases confidence
```

### Measurable Metrics
The system provides quantifiable metrics:

#### Recommendation Quality
- **Genre Match Rate**: Percentage of recommendations in preferred genres
- **Exclusion Accuracy**: Percentage of recommendations not already rated/watchlisted
- **Score Distribution**: Analysis of recommendation scores

#### User Engagement
- **Click-through Rate**: User interaction with recommendations
- **Rating Conversion**: Users rating recommended movies
- **Watchlist Addition**: Users adding recommendations to watchlist

#### Algorithm Performance
- **Execution Time**: Time to generate recommendations
- **Database Queries**: Number of database operations
- **Memory Usage**: Resource consumption during processing

### Testability
The system supports comprehensive testing:

#### Unit Testing
- Individual algorithm components testable
- Mock data for consistent test scenarios
- Edge case validation (no ratings, new users)

#### Integration Testing
- End-to-end recommendation flow
- Database interaction validation
- API endpoint testing

#### Performance Testing
- Load testing with multiple users
- Scalability assessment
- Resource usage monitoring

## Implementation Details

### Service Layer Architecture

#### RecommendationService
```go
type RecommendationService struct {
    movieRepo         *repositories.MovieRepository
    ratingRepo        *repositories.RatingRepository
    watchlistRepo     *repositories.WatchlistRepository
    recommendationRepo *repositories.RecommendationRepository
}
```

#### Method Signatures
```go
func (s *RecommendationService) GetRecommendations(userID primitive.ObjectID, limit int) ([]*models.Movie, error)
func (s *RecommendationService) GetDetailedRecommendations(userID primitive.ObjectID, limit int) ([]*Recommendation, error)
```

### Repository Layer Integration

#### High-Rated Genres Query
```go
func (r *RecommendationRepository) GetHighRatedGenres(userID primitive.ObjectID, threshold int) ([]string, error)
```

**MongoDB Aggregation**:
```javascript
[
  // Match user's high ratings
  { "$match": { "user_id": userID, "rating": { "$gte": threshold } } },
  
  // Lookup movie details
  { "$lookup": { "from": "movies", "localField": "movie_id", "foreignField": "_id", "as": "movie" } },
  
  // Unwind movie array
  { "$unwind": "$movie" },
  
  // Split genre string into array
  { "$project": { "genres": { "$split": ["$movie.genre", ","] } } },
  
  // Unwind genres
  { "$unwind": "$genres" },
  
  // Trim whitespace
  { "$project": { "genre": { "$trim": { "input": "$genres" } } } },
  
  // Filter empty genres
  { "$match": { "genre": { "$ne": "" } } },
  
  // Group and count
  { "$group": { "_id": "$genre", "count": { "$sum": 1 } } },
  
  // Sort by count
  { "$sort": { "count": -1 } }
]
```

#### Genre-Based Movie Search
```go
func (r *RecommendationRepository) GetMoviesByGenres(genres []string, excludeIDs []primitive.ObjectID, limit int) ([]*models.Movie, error)
```

**Query Logic**:
- Regex pattern matching for multiple genres
- Exclusion of specified movie IDs
- Limit and sort for consistent results

### Data Flow Architecture

#### Request Processing
```
User Request → JWT Validation → Recommendation Service → Repository Queries → Response Generation
```

#### Database Operations
```
1. Query user ratings (ratings collection)
2. Query user watchlist (watchlists collection)
3. Aggregate genre preferences (ratings + movies collections)
4. Search candidate movies (movies collection)
5. Apply exclusions and limits (movies collection)
```

#### Response Construction
```
1. Calculate recommendation scores
2. Assign confidence levels
3. Generate recommendation reasons
4. Format response data
5. Return to client
```

## Performance Characteristics

### Computational Complexity
- **Time Complexity**: O(n + m) where n = user ratings, m = candidate movies
- **Space Complexity**: O(k) where k = number of recommendations
- **Database Queries**: 3-4 optimized queries per recommendation

### Scalability Considerations
- **User Growth**: Linear scaling with user base
- **Movie Growth**: Minimal impact (indexed searches)
- **Rating Growth**: Improved recommendations with more data

### Optimization Strategies
- **Database Indexing**: Composite indexes for frequent queries
- **Query Optimization**: Efficient MongoDB aggregations
- **Caching**: Movie data cached from OMDb API
- **Pagination**: Limit results for large datasets

## Limitations and Considerations

### Current Limitations
- **Cold Start Problem**: New users get generic recommendations
- **Data Sparsity**: Limited recommendations with few ratings
- **Genre Granularity**: Simple genre matching without semantic analysis
- **Static Preferences**: No temporal preference evolution

### Mitigation Strategies
- **Popular Movie Fallback**: Ensures recommendations for all users
- **Minimum Rating Threshold**: Requires minimum ratings for personalization
- **Genre Weighting**: Balances frequency and rating quality
- **Hybrid Approach**: Combines multiple recommendation strategies

### Future Enhancement Opportunities
- **Collaborative Filtering**: Similar user preferences
- **Content-Based Filtering**: Movie attribute analysis
- **Temporal Dynamics**: Time-weighted preferences
- **Social Features**: Friend recommendations
- **Machine Learning**: Advanced pattern recognition

## Evaluation Metrics

### Accuracy Metrics
- **Precision**: Percentage of relevant recommendations
- **Recall**: Coverage of user's preferences
- **F1-Score**: Balance of precision and recall

### Diversity Metrics
- **Genre Diversity**: Variety in recommended genres
- **Year Distribution**: Range of movie release years
- **Rating Distribution**: Spread of movie ratings

### Engagement Metrics
- **Click-Through Rate**: User interaction with recommendations
- **Conversion Rate**: Users acting on recommendations
- **Satisfaction**: User feedback on recommendation quality

### System Metrics
- **Response Time**: Recommendation generation speed
- **Throughput**: Recommendations per second
- **Resource Usage**: CPU and memory consumption

## Conclusion

The rule-based recommendation system provides a solid foundation for movie suggestions while maintaining transparency and determinism. The algorithm is well-suited for academic evaluation due to its explainable nature and consistent behavior. The system balances simplicity with effectiveness, providing meaningful recommendations based on user preferences while ensuring good performance and scalability.

The modular architecture allows for future enhancements and integration with more sophisticated techniques while maintaining the core rule-based approach that makes the system suitable for both production use and academic study.
