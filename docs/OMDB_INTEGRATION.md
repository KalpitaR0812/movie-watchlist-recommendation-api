# OMDb Integration Documentation

## Overview

This document explains the integration with the OMDb (Open Movie Database) API for fetching movie data. The system implements a two-tier approach: search for movies using basic information, then fetch complete details for caching and recommendation purposes.

## OMDb API Usage

### Search API Endpoint

**Purpose**: Find movies by title or partial title matches.

**Endpoint**: `http://www.omdbapi.com/?apikey={API_KEY}&s={QUERY}`

**Method**: HTTP GET

**Parameters**:
- `apikey`: OMDb API authentication key
- `s`: Search query (movie title or partial title)

**Response Format**:
```json
{
  "Search": [
    {
      "Title": "Movie Title",
      "Year": "2023",
      "imdbID": "tt1234567",
      "Type": "movie",
      "Poster": "https://example.com/poster.jpg"
    }
  ],
  "totalResults": "1",
  "Response": "True"
}
```

**Implementation Details**:
- Used in `SearchMovies()` service method
- Returns basic movie information for search results
- Limited fields: Title, Year, IMDb ID, Type, Poster
- Does not include genre, director, plot, or rating details

### Movie Details API Endpoint

**Purpose**: Fetch complete movie information including all details needed for recommendations.

**Endpoint**: `http://www.omdbapi.com/?apikey={API_KEY}&i={IMDB_ID}`

**Method**: HTTP GET

**Parameters**:
- `apikey`: OMDb API authentication key
- `i`: IMDb ID (exact match)

**Response Format**:
```json
{
  "Title": "Movie Title",
  "Year": "2023",
  "Rated": "PG-13",
  "Released": "01 Jan 2023",
  "Runtime": "120 min",
  "Genre": "Action, Sci-Fi",
  "Director": "Director Name",
  "Writer": "Writer Name",
  "Actors": "Actor 1, Actor 2",
  "Plot": "Movie plot description...",
  "Language": "English",
  "Country": "USA",
  "Awards": "Awards information",
  "Poster": "https://example.com/poster.jpg",
  "Ratings": [
    {
      "Source": "Internet Movie Database",
      "Value": "7.5/10"
    }
  ],
  "Metascore": "75",
  "imdbRating": "7.5",
  "imdbVotes": "100,000",
  "imdbID": "tt1234567",
  "Type": "movie",
  "DVD": "01 Mar 2023",
  "BoxOffice": "$100,000,000",
  "Production": "Production Company",
  "Website": "https://example.com",
  "Response": "True"
}
```

**Implementation Details**:
- Used in `GetMovieDetails()` and `GetOrCreateByIMDbID()` service methods
- Returns comprehensive movie information
- Includes genre, director, plot, and rating details
- Essential data for recommendation engine functionality

## Two-Tier Caching Strategy

### Tier 1: Search Results
- **Trigger**: User searches for movies by title
- **API Call**: OMDb Search API (`?s=` parameter)
- **Data Received**: Basic movie information (Title, Year, IMDb ID, Poster)
- **Caching**: No immediate caching of search results
- **Purpose**: Provide search results to user for selection

### Tier 2: Full Details Caching
- **Trigger**: User selects a specific movie or requests movie by IMDb ID
- **API Call**: OMDb Details API (`?i=` parameter)
- **Data Received**: Complete movie information including genre, director, plot, ratings
- **Caching**: Full movie details stored in MongoDB movies collection
- **Purpose**: Enable recommendation engine and reduce future API calls

## Caching Implementation

### Cache Check Logic
```go
// Check if movie exists in cache first
if movie, err := s.movieRepo.FindByIMDbID(imdbID); err == nil {
    return movie, nil  // Return cached data
}

// Fetch from OMDb if not cached
omdbMovie, err := s.fetchFromOMDb(imdbID)
if err != nil {
    return nil, err
}

// Cache the complete movie data
s.movieRepo.Create(omdbMovie)
return omdbMovie, nil
```

### Cache Storage
- **Collection**: `movies` in MongoDB
- **Index**: Unique index on `imdb_id` field
- **Timestamp**: `cached_at` field tracks when data was fetched
- **Update Strategy**: Movies are cached once and never updated (current implementation)

## Duplicate API Call Avoidance

### Search Flow Optimization
1. **User Search**: Search API called with user query
2. **Result Processing**: For each search result:
   - Check if movie already exists in cache
   - If cached, skip details fetch
   - If not cached, fetch full details and cache
3. **Response**: Return search results to user

### Individual Movie Request Flow
1. **Movie Request**: Check cache first by IMDb ID
2. **Cache Hit**: Return cached movie data immediately
3. **Cache Miss**: Fetch from OMDb Details API, cache result, return data

### Benefits of Caching Strategy
- **Reduced API Calls**: Each movie is fetched from OMDb only once
- **Improved Performance**: Cached data retrieval is faster than external API calls
- **Reliability**: System continues functioning during OMDb API outages (for cached movies)
- **Cost Efficiency**: Minimizes usage of OMDb API quotas
- **Recommendation Support**: Cached genre data enables recommendation engine

## Error Handling

### OMDb API Errors
- **Invalid API Key**: Returns configuration error
- **Movie Not Found**: Returns appropriate error to user
- **API Rate Limits**: Implements retry logic with exponential backoff
- **Network Issues**: Returns connection errors with user-friendly messages

### Cache Errors
- **Database Connection**: Fallback to direct API calls if cache unavailable
- **Data Corruption**: Re-fetch from OMDb if cached data is invalid
- **Index Issues**: Automatic index creation on application startup

## Performance Considerations

### API Rate Limiting
- OMDb API has rate limits (approximately 1,000 requests per day for free tier)
- Caching strategy minimizes API calls to stay within limits
- Batch processing reduces individual API requests

### Response Time Optimization
- Cache lookup is significantly faster than external API calls
- MongoDB indexes ensure fast cache retrieval
- Concurrent processing for multiple movie details fetching

### Data Freshness
- Movie data is relatively static (basic movie information doesn't change frequently)
- Current implementation caches indefinitely
- Future enhancement could implement cache expiration for very old data

## Security Considerations

### API Key Management
- OMDb API key stored in environment variables
- Key is not exposed in client responses
- Key validation on application startup

### Data Validation
- All OMDb responses are validated before caching
- Required fields (title, IMDb ID, genre) are validated
- Malformed responses are rejected with appropriate error messages

## Integration Points

### Movie Service Integration
- `SearchMovies()`: Uses search API for user queries
- `GetMovieDetails()`: Uses details API for individual movie requests
- `GetOrCreateByIMDbID()`: Uses details API with caching logic

### Repository Integration
- `FindByIMDbID()`: Checks cache before API calls
- `Create()`: Stores fetched movie data in cache
- MongoDB handles persistence and indexing

### Recommendation Engine Integration
- Cached genre data is essential for recommendation logic
- Recommendation engine relies on cached movie details
- No external API calls needed for recommendation generation

## Monitoring and Logging

### API Call Tracking
- All OMDb API calls are logged with timestamps
- Cache hit/miss ratios are tracked
- Error rates are monitored for API reliability

### Cache Performance
- Cache lookup times are monitored
- Database query performance is tracked
- Cache size and growth patterns are analyzed

## Future Enhancements

### Cache Refresh Strategy
- Implement periodic cache refresh for old movie data
- Update ratings and metadata for cached movies
- Maintain data freshness while minimizing API calls

### Advanced Caching
- Implement Redis caching layer for frequently accessed movies
- Add cache warming for popular movies
- Implement cache invalidation strategies

### API Optimization
- Implement request batching for multiple movie details
- Add support for OMDb API advanced features
- Optimize search result processing
