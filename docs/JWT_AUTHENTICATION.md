# JWT Authentication Documentation

## Overview

The Movie Watchlist system implements JSON Web Token (JWT) based authentication to secure API endpoints and manage user sessions. The authentication system provides stateless authentication with token-based authorization, ensuring secure access to user-specific resources.

## JWT Implementation Architecture

### Authentication Flow
```
User Login → Credential Validation → JWT Generation → Token Return → Client Storage → Token Validation → Authorized Access
```

### Token Components
- **Header**: Algorithm and token type information
- **Payload**: User claims and metadata
- **Signature**: Cryptographic signature for integrity verification

## JWT Token Generation

### Claims Structure
```go
type Claims struct {
    UserID   primitive.ObjectID `json:"user_id"`
    Username string            `json:"username"`
    Email    string            `json:"email"`
    jwt.RegisteredClaims
}
```

### Token Generation Process
```go
func (m *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
    // Validate user ID
    if user.ID.IsZero() {
        return "", fmt.Errorf("invalid user ID")
    }

    // Create claims with user information
    claims := &Claims{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "movie-watchlist-api",
            Subject:   user.ID.Hex(),
        },
    }

    // Create token with signing method
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    // Sign token with secret key
    tokenString, err := token.SignedString([]byte(m.jwtSecret))
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %w", err)
    }

    return tokenString, nil
}
```

### Token Configuration
- **Algorithm**: HS256 (HMAC with SHA-256)
- **Expiration**: 24 hours from issuance
- **Issuer**: "movie-watchlist-api"
- **Subject**: User's MongoDB ObjectID (hex string)

### Token Payload Example
```json
{
  "user_id": "507f1f77bcf86cd799439011",
  "username": "john_doe",
  "email": "john@example.com",
  "exp": 1703980800,
  "iat": 1703894400,
  "nbf": 1703894400,
  "iss": "movie-watchlist-api",
  "sub": "507f1f77bcf86cd799439011"
}
```

## Bearer Token Extraction

### HTTP Header Format
```
Authorization: Bearer <jwt_token>
```

### Token Extraction Logic
```go
func extractToken(authHeader string) string {
    if authHeader == "" {
        return ""
    }

    // Remove "Bearer " prefix if present
    if strings.HasPrefix(authHeader, "Bearer ") {
        return authHeader[7:]
    }

    return authHeader
}
```

### Extraction Process
1. **Header Validation**: Check for Authorization header presence
2. **Prefix Removal**: Strip "Bearer " prefix if present
3. **Token Isolation**: Extract pure JWT token string
4. **Format Validation**: Ensure token follows JWT structure

### Error Handling
- **Missing Header**: Return 401 Unauthorized
- **Invalid Format**: Return 401 Unauthorized
- **Empty Token**: Return 401 Unauthorized

## Token Validation Middleware

### Middleware Implementation
```go
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header is required",
                "code":  "AUTH_HEADER_MISSING",
            })
            c.Abort()
            return
        }

        // Validate token
        claims, err := m.ValidateToken(authHeader)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid or expired token",
                "code":  "INVALID_TOKEN",
                "details": err.Error(),
            })
            c.Abort()
            return
        }

        // Set user information in context
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("email", claims.Email)
        c.Set("claims", claims)

        c.Next()
    }
}
```

### Validation Steps
1. **Header Extraction**: Get Authorization header from request
2. **Token Parsing**: Parse JWT token with claims
3. **Signature Verification**: Verify token signature with secret key
4. **Claims Validation**: Validate token claims and expiration
5. **Context Injection**: Set user information in request context

### Token Parsing Logic
```go
func (m *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
    // Parse token with claims
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(m.jwtSecret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }

    // Extract and validate claims
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        // Additional validation
        if claims.UserID.IsZero() {
            return nil, fmt.Errorf("invalid user ID in token")
        }

        // Check expiration
        if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
            return nil, fmt.Errorf("token has expired")
        }

        return claims, nil
    }

    return nil, fmt.Errorf("invalid token")
}
```

## User Context Injection

### Context Variables
After successful validation, the middleware injects user information into the Gin context:

```go
c.Set("user_id", claims.UserID)
c.Set("username", claims.Username)
c.Set("email", claims.Email)
c.Set("claims", claims)
```

### Context Access in Handlers
```go
func (h *MovieHandler) GetMovie(c *gin.Context) {
    // Get user ID from context
    userIDValue, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }

    userID, ok := userIDValue.(primitive.ObjectID)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
        return
    }

    // Use userID for user-specific operations
    movie, err := h.movieService.GetMovieByID(userID, movieID)
    // ... rest of handler logic
}
```

### Helper Functions for Context Access
```go
// GetCurrentUser retrieves complete claims from context
func (m *AuthMiddleware) GetCurrentUser(c *gin.Context) (*Claims, bool) {
    claims, exists := c.Get("claims")
    if !exists {
        return nil, false
    }

    userClaims, ok := claims.(*Claims)
    if !ok {
        return nil, false
    }

    return userClaims, true
}

// GetUserID retrieves user ID from context
func (m *AuthMiddleware) GetUserID(c *gin.Context) (primitive.ObjectID, bool) {
    userIDValue, exists := c.Get("user_id")
    if !exists {
        return primitive.NilObjectID, false
    }

    userID, ok := userIDValue.(primitive.ObjectID)
    if !ok {
        return primitive.NilObjectID, false
    }

    return userID, true
}
```

## Security Considerations

### Secret Key Management
- **Environment Variables**: JWT secret stored in `JWT_SECRET` environment variable
- **Strong Keys**: Minimum 32-character recommended secret
- **Key Rotation**: Periodic secret key rotation for enhanced security
- **Access Control**: Limited access to secret key configuration

### Token Security Features
- **HMAC-SHA256**: Cryptographically secure signing algorithm
- **Time-Limited**: 24-hour expiration prevents long-term token abuse
- **Issuer Validation**: Ensures tokens from correct application
- **Subject Validation**: Links token to specific user

### Common Security Vulnerabilities Prevented
- **Algorithm Confusion**: Explicit algorithm validation prevents downgrade attacks
- **Token Tampering**: Signature verification prevents content modification
- **Replay Attacks**: Expiration and "not before" claims prevent replay
- **Cross-Site Request Forgery**: Token-based authentication prevents CSRF

## Error Handling and Response Formats

### Authentication Error Responses
```json
{
  "error": "Authorization header is required",
  "code": "AUTH_HEADER_MISSING"
}
```

```json
{
  "error": "Invalid or expired token",
  "code": "INVALID_TOKEN",
  "details": "token has expired"
}
```

### Error Categories
- **AUTH_HEADER_MISSING**: No Authorization header provided
- **INVALID_TOKEN**: Token parsing or validation failed
- **TOKEN_EXPIRED**: Token expiration time passed
- **INVALID_USER_ID**: User ID in token is invalid

### HTTP Status Codes
- **401 Unauthorized**: Authentication failures
- **403 Forbidden**: Authorization failures (future use)
- **500 Internal Server Error**: System errors during validation

## Token Refresh Strategy

### Current Implementation
The system currently implements a simple 24-hour token expiration without automatic refresh. Users must re-authenticate after token expiration.

### Refresh Token Logic (Future Enhancement)
```go
func (m *AuthMiddleware) RefreshToken(c *gin.Context) (string, error) {
    claims, exists := m.GetCurrentUser(c)
    if !exists {
        return "", fmt.Errorf("user not authenticated")
    }

    // Create new claims with same user info but new expiration
    newClaims := &Claims{
        UserID:   claims.UserID,
        Username: claims.Username,
        Email:    claims.Email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "movie-watchlist-api",
            Subject:   claims.UserID.Hex(),
        },
    }

    // Create and sign new token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
    tokenString, err := token.SignedString([]byte(m.jwtSecret))
    if err != nil {
        return "", fmt.Errorf("failed to sign new token: %w", err)
    }

    return tokenString, nil
}
```

## Optional Authentication

### Optional Auth Middleware
```go
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            claims, err := m.ValidateToken(authHeader)
            if err == nil {
                c.Set("user_id", claims.UserID)
                c.Set("username", claims.Username)
                c.Set("email", claims.Email)
                c.Set("claims", claims)
            }
            // Continue without authentication if token invalid
        }
        c.Next()
    }
}
```

### Use Cases
- **Public Endpoints**: Movie search without authentication
- **Enhanced Responses**: Additional data for authenticated users
- **Gradual Authentication**: Basic features without login, enhanced with login

## Performance Considerations

### Token Validation Overhead
- **CPU Usage**: Minimal HMAC-SHA256 computation
- **Memory Usage**: Small token strings in memory
- **Network Overhead**: Authorization header (~500 bytes)
- **Database Queries**: No database queries for token validation

### Optimization Strategies
- **Token Caching**: Cache validated tokens (security trade-off)
- **Async Validation**: Background token refresh
- **Connection Pooling**: Reuse HTTP connections
- **Compression**: Token compression for large payloads

### Monitoring Metrics
- **Validation Time**: Time to validate tokens
- **Failure Rate**: Percentage of invalid tokens
- **Expiration Rate**: Tokens expiring vs refreshing
- **Security Events**: Suspicious token patterns

## Integration Points

### Authentication Middleware Application
```go
// Apply to all API routes
api.Use(middleware.RequireAuth())

// Apply to specific route groups
admin.Use(middleware.RequireRole("admin"))

// Optional authentication for public endpoints
public.GET("/movies/search", middleware.OptionalAuth(), movieHandler.SearchMovies)
```

### Handler Integration
```go
type MovieHandler struct {
    movieService *services.MovieService
}

func (h *MovieHandler) GetUserRecommendations(c *gin.Context) {
    // Extract user from context
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
        return
    }

    // Use user ID for personalized recommendations
    recommendations, err := h.movieService.GetRecommendations(userID.(primitive.ObjectID))
    // ... rest of implementation
}
```

### Service Layer Integration
```go
type UserService struct {
    userRepo *repositories.UserRepository
    jwtSecret string
}

func (s *UserService) Login(email, password string) (*models.User, string, error) {
    // Validate credentials
    user, err := s.userRepo.FindByEmail(email)
    if err != nil || !bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
        return nil, "", fmt.Errorf("invalid credentials")
    }

    // Generate JWT token
    token, err := s.generateJWTToken(user)
    if err != nil {
        return nil, "", err
    }

    return user, token, nil
}
```

## Testing Considerations

### Unit Testing
```go
func TestJWTGeneration(t *testing.T) {
    middleware := NewAuthMiddleware("test-secret")
    user := &models.User{
        ID:       primitive.NewObjectID(),
        Username: "testuser",
        Email:    "test@example.com",
    }

    token, err := middleware.GenerateToken(user)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // Validate token
    claims, err := middleware.ValidateToken(token)
    assert.NoError(t, err)
    assert.Equal(t, user.ID, claims.UserID)
    assert.Equal(t, user.Username, claims.Username)
    assert.Equal(t, user.Email, claims.Email)
}
```

### Integration Testing
```go
func TestAuthenticationMiddleware(t *testing.T) {
    router := gin.New()
    middleware := NewAuthMiddleware("test-secret")
    
    router.Use(middleware.RequireAuth())
    router.GET("/protected", func(c *gin.Context) {
        userID, _ := middleware.GetUserID(c)
        c.JSON(http.StatusOK, gin.H{"user_id": userID.Hex()})
    })

    // Test with valid token
    token := generateTestToken()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)
    
    assert.Equal(t, 200, resp.Code)
}
```

## Best Practices

### Token Management
- **Short Expiration**: 24-hour balance between security and usability
- **Secure Storage**: Store tokens securely on client side
- **HTTPS Only**: Transmit tokens only over HTTPS connections
- **Token Revocation**: Implement token blacklist for compromised tokens

### Error Handling
- **Consistent Responses**: Standardized error response format
- **Security Logging**: Log authentication failures for monitoring
- **Rate Limiting**: Prevent brute force attacks on authentication
- **User Feedback**: Clear error messages for users

### Configuration Management
- **Environment Variables**: Store secrets in environment, not code
- **Configuration Validation**: Validate JWT configuration on startup
- **Secret Rotation**: Plan for periodic secret key changes
- **Backup Strategy**: Secure backup of authentication configuration

## Future Enhancements

### Advanced Security Features
- **Token Blacklisting**: Immediate token revocation capability
- **Multi-Factor Authentication**: Additional security layers
- **Device Fingerprinting**: Track token usage patterns
- **Anomaly Detection**: Identify suspicious authentication patterns

### Performance Optimizations
- **Token Compression**: Reduce token size for mobile clients
- **Batch Validation**: Validate multiple tokens efficiently
- **Caching Layer**: Cache frequently validated tokens
- **Async Processing**: Background token validation

### User Experience Improvements
- **Refresh Tokens**: Seamless token renewal
- **Remember Me**: Extended session options
- **Multi-Device Support**: Manage sessions across devices
- **Social Login Integration**: Third-party authentication options

## Conclusion

The JWT authentication system provides a robust, secure, and scalable solution for user authentication in the Movie Watchlist application. The implementation balances security requirements with usability considerations, ensuring protected access to user-specific resources while maintaining good performance and developer experience.

The stateless nature of JWT tokens enables horizontal scalability and simplified session management, while the comprehensive validation and error handling ensure system security and reliability. The modular architecture allows for future enhancements and integration with advanced security features as the application evolves.
