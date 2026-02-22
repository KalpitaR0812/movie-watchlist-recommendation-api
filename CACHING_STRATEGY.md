# Caching Strategy

## Introduction

The Movie Watchlist & Recommendation API implements a local database caching strategy for movie data retrieved from the OMDb API. This approach minimizes external API dependencies, improves response times, and reduces operational costs while maintaining data consistency and service reliability.

## Problem Statement

External API dependencies introduce several challenges for a production application:

- **Rate Limitations**: OMDb API free tier limits requests to 1000 calls per day
- **Performance Variability**: External API response times range from 500-1000ms
- **Service Dependency**: Application functionality is tied to external service availability
- **Cost Implications**: Paid API tiers become necessary at scale
- **Reliability Risks**: Network issues and service outages affect application uptime

## Caching Approach

### Cache Storage

The system utilizes PostgreSQL as the primary cache storage medium. Movie data is stored in the `movies` table with the following structure:

- **Cache Key**: IMDb ID (unique identifier for each movie)
- **Cache Entry**: Complete movie metadata from OMDb API response
- **Cache Metadata**: Timestamp indicating when data was fetched from external API

### Cache Identification

Each movie is uniquely identified by its IMDb ID, which serves as the cache key. This approach ensures that duplicate requests for the same movie are served from the local cache without additional API calls.

## Cache Population Strategy

### Movie Search Operations

Movie search operations bypass caching to ensure fresh results:

```
Client Request → API → OMDb API → Response
```

Search results return basic movie information (title, year, IMDb ID) but do not include complete details. This approach ensures users receive current search results while allowing the system to cache individual movies as they are accessed.

### Movie Detail Operations

Movie detail operations implement a cache-first strategy:

```
Client Request → Check Local Cache
                ↓
          [Cache Hit?] → Yes → Return Cached Data
                ↓
                No → OMDb API → Store in Cache → Return Data
```

When a movie detail request is received, the system first checks if the movie exists in the local database. If found, the cached data is returned immediately. If not found, the system fetches the data from OMDb API and stores it locally for future requests.

## Implementation Details

### Cache Check Logic

The caching logic is implemented in the movie service layer:

```go
func (s *MovieService) GetMovieDetails(imdbID string) (*models.Movie, error) {
    // Check cache first
    if movie, err := s.movieRepo.FindByIMDbID(imdbID); err == nil {
        return movie, nil // Cache hit
    }
    
    // Cache miss - fetch from external API
    movie, err := s.fetchFromOMDbAPI(imdbID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache for future requests
    if err := s.movieRepo.Create(movie); err != nil {
        return nil, err
    }
    
    return movie, nil
}
```

### Cache Data Structure

Cached movie data includes all relevant OMDb API fields:

- **Identifiers**: IMDb ID, title, year
- **Classification**: Genre, director, runtime
- **Content**: Plot summary, poster URL
- **Ratings**: IMDb community rating
- **Metadata**: Cache timestamp, creation/update times

### Cache Metadata

Each cached movie entry includes metadata to support cache management:

- **Cached At**: Timestamp when movie was first fetched from OMDb API
- **Created At**: Database record creation time
- **Updated At**: Last modification time
- **Deleted At**: Soft delete support for cache invalidation

## Cache Consistency and Validity

### Data Immutability

Movie metadata is treated as effectively immutable:

- **Directors**: Do not change
- **Release Years**: Remain constant
- **Basic Facts**: Title, runtime, genre classifications
- **Ratings**: May change slowly but not critical for functionality

### Cache Lifetime

The system implements a persistent cache with no automatic expiration:

- **Rationale**: Movie metadata changes infrequently
- **Benefit**: Eliminates cache invalidation complexity
- **Strategy**: Manual refresh for specific movies if needed
- **Acceptable**: Movie metadata changes infrequently enough to justify approach

### Data Freshness

While the cache is persistent, the system maintains data freshness through:

- **Organic Growth**: New movies are added as users search for them
- **User-Driven Updates**: Popular movies are cached through regular usage
- **Manual Refresh**: Administrative functionality to update specific movies

## Performance and Reliability Benefits

### Response Time Improvements

Caching provides significant performance improvements:

- **Cache Hit**: 5-10ms (database query)
- **Cache Miss**: 500-1000ms (API call + database write)
- **Improvement Factor**: 50-200x faster for cached data

### External API Call Reduction

The caching strategy reduces external API dependency:

- **Initial User**: 20-30 API calls (first-time searches)
- **Active User**: 5-10 API calls per day (new searches)
- **Power User**: 1-2 API calls per day (rare new content)
- **Reduction**: 80-95% fewer API calls compared to direct API usage

### Service Availability

Local caching enhances service reliability:

- **External API Down**: Service remains functional for cached content
- **Network Issues**: No impact on cached movie access
- **Rate Limiting**: Reduced exposure to API rate limits
- **Consistent Performance**: Predictable response times for cached data

## Trade-offs and Limitations

### Storage Overhead

Database storage requirements for caching:

- **Per Movie**: Approximately 2KB including all metadata
- **10,000 Movies**: ~20MB database storage
- **100,000 Movies**: ~200MB database storage
- **Cost Efficiency**: Minimal compared to API subscription costs

### Cache Invalidation

The current strategy does not implement automatic cache invalidation:

- **Assumption**: Movie metadata remains valid indefinitely
- **Trade-off**: Simplified implementation at cost of potential staleness
- **Mitigation**: Manual refresh capability for critical updates
- **Acceptable**: Movie metadata changes infrequently enough to justify approach

### Memory Usage

PostgreSQL manages memory efficiently:

- **Connection Pooling**: Reuses database connections
- **Query Optimization**: Indexes on frequently accessed fields
- **Resource Management**: Database handles memory allocation
- **Scalability**: Can handle millions of cached records

## Monitoring and Metrics

### Cache Performance Metrics

Key metrics to monitor cache effectiveness:

- **Cache Hit Rate**: Percentage of requests served from local cache
- **Cache Miss Rate**: Percentage requiring external API calls
- **Average Response Time**: Compare cached vs uncached requests
- **API Call Volume**: Daily and monthly external API usage
- **Cache Size**: Number of movies stored locally

### Performance Analysis

Sample monitoring query:

```sql
SELECT 
    COUNT(*) as total_movies,
    COUNT(cached_at) as cached_movies,
    ROUND(COUNT(cached_at) * 100.0 / COUNT(*), 2) as cache_hit_rate
FROM movies;
```

### Operational Monitoring

Database-level metrics provide insights into cache utilization:

- **Storage Growth**: Track cache size over time
- **Access Patterns**: Identify most frequently accessed movies
- **Performance Trends**: Monitor query execution times
- **Error Rates**: Track cache-related failures

## Conclusion

The implemented caching strategy provides optimal balance between performance, cost efficiency, and reliability for the Movie Watchlist & Recommendation API. By leveraging PostgreSQL as a persistent cache, the system achieves:

- **Performance**: 50-200x faster response times for cached content
- **Cost Reduction**: 80-95% fewer external API calls
- **Reliability**: Service independence from external API limitations
- **Scalability**: Efficient handling of growing movie database

This approach is particularly suitable for the application's requirements because movie metadata is relatively static, user access patterns follow predictable distributions, and the cost of database storage is minimal compared to external API subscription costs.
