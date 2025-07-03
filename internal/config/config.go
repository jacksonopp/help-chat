package config

import (
	"os"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Host string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	FilePath string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SecretKey       string
	AccessTokenTTL  string
	RefreshTokenTTL string
	Issuer          string
	// Cookie configuration
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite string
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Host: getEnv("HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			FilePath: getEnv("DB_FILE", "helpchat.db"),
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			AccessTokenTTL:  getEnv("JWT_ACCESS_TOKEN_TTL", "15m"),
			RefreshTokenTTL: getEnv("JWT_REFRESH_TOKEN_TTL", "7d"),
			Issuer:          getEnv("JWT_ISSUER", "helpchat"),
			CookieDomain:    getEnv("JWT_COOKIE_DOMAIN", ""),
			CookieSecure:    getEnv("JWT_COOKIE_SECURE", "false") == "true",
			CookieSameSite:  getEnv("JWT_COOKIE_SAME_SITE", "Lax"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getCORSOrigins(),
			AllowedMethods:   []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "content-type"},
			AllowCredentials: true,
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getCORSOrigins gets CORS origins from environment variable or returns default values
func getCORSOrigins() []string {
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		// Split by comma and trim whitespace
		originList := strings.Split(origins, ",")
		for i, origin := range originList {
			originList[i] = strings.TrimSpace(origin)
		}
		return originList
	}

	// Default origins for development
	return []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:5173",
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:4173", // Vite preview
		"http://localhost:4000", // Common dev port
		"http://localhost:4200", // Angular default
		"http://localhost:4300", // Additional Angular port
	}
}
