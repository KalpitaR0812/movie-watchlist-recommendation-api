package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	OMDbAPIKey  string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "mongodb://localhost:27017/movie_watchlist"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		OMDbAPIKey:  getEnv("OMDB_API_KEY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
