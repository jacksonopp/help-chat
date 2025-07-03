package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
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
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
