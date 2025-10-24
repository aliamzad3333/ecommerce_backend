package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URI         string
	Database    string
	Timeout     time.Duration
	MaxPoolSize uint64
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			URI:         getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database:    getEnv("DATABASE_NAME", "ecommerce"),
			Timeout:     getDurationEnv("DB_TIMEOUT", 10*time.Second),
			MaxPoolSize: getUint64Env("DB_MAX_POOL_SIZE", 100),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "default-secret-key"),
			Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getUint64Env(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}
