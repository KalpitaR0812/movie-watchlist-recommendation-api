package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Claims struct {
	UserID primitive.ObjectID `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Step 2: Validate Bearer token format
		tokenString, err := extractBearerToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "INVALID_TOKEN_FORMAT",
			})
			c.Abort()
			return
		}

		// Step 3: Parse and validate JWT token
		claims, err := parseAndValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Step 4: Inject user_id into request context
		c.Set("user_id", claims.UserID)
		c.Set("user_claims", claims)
		
		// Step 5: Continue to next handler
		c.Next()
	}
}

// extractBearerToken extracts the Bearer token from the Authorization header
func extractBearerToken(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", fmt.Errorf("authorization header must be in format 'Bearer <token>'")
	}
	
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}
	
	return token, nil
}

// parseAndValidateToken parses and validates the JWT token
func parseAndValidateToken(tokenString, jwtSecret string) (*Claims, error) {
	claims := &Claims{}
	
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		// Return the secret key for validation
		return []byte(jwtSecret), nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	// Additional validation: check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}
	
	// Additional validation: check issued at
	if claims.IssuedAt != nil && claims.IssuedAt.Time.After(time.Now().Add(5*time.Minute)) {
		return nil, fmt.Errorf("token issued in the future")
	}
	
	return claims, nil
}

// GenerateToken generates a JWT token for the given user ID
func GenerateToken(userID primitive.ObjectID, jwtSecret string) (string, error) {
	if userID.IsZero() {
		return "", fmt.Errorf("user ID cannot be empty")
	}
	
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT secret cannot be empty")
	}
	
	// Create claims with expiration and issued at
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "movie-watchlist-api",
			Subject:   userID.Hex(),
		},
	}
	
	// Create token with signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, jwtSecret string) (*Claims, error) {
	return parseAndValidateToken(tokenString, jwtSecret)
}

// RefreshToken generates a new token with extended expiration
func RefreshToken(oldTokenString, jwtSecret string) (string, error) {
	// Parse old token
	claims, err := parseAndValidateToken(oldTokenString, jwtSecret)
	if err != nil {
		return "", fmt.Errorf("invalid token for refresh: %w", err)
	}
	
	// Generate new token with same user ID
	return GenerateToken(claims.UserID, jwtSecret)
}
