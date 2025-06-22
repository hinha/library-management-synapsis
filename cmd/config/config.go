package config

import (
	"github.com/joho/godotenv"
	"os"
	"time"
)

var _ = godotenv.Load(".env")

// GetEnv retrieves an environment variable or returns a default value if not found
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ServiceConfig holds configuration for a microservice
type ServiceConfig struct {
	// Database configuration
	DbHost     string
	DbUser     string
	DbPassword string
	DbName     string
	DbPort     string

	// Cache configuration
	CacheHost     string
	CachePort     string
	CachePassword string
	CacheDbToken  string

	// Server configuration
	GrpcAddr string
	HttpAddr string

	// JWT configuration (for user service)
	JwtSecret string
}

// JWT_TOKEN_EXPIRATION=24h
var (
	LogDebug              = GetEnv("LOG_DEBUG", "false") == "true"
	InitialAdminEmail     = GetEnv("INITIAL_ADMIN_EMAIL", "")
	InitialAdminPassword  = GetEnv("INITIAL_ADMIN_PASSWORD", "")
	JwtTokenExpiration, _ = time.ParseDuration(GetEnv("JWT_TOKEN_EXPIRATION", "24h"))

	RedisKeyUserPrefix = GetEnv("REDIS_KEY_USER_PREFIX", "user:")

	SharedGrpcAuthServiceAddr = GetEnv("USER_GRPC_ADDR", ":50051")
)

// LoadUserServiceConfig loads configuration for the user service
func LoadUserServiceConfig() ServiceConfig {
	return ServiceConfig{
		DbHost:        GetEnv("USER_DB_HOST", "localhost"),
		DbUser:        GetEnv("USER_DB_USER", "postgres"),
		DbPassword:    GetEnv("USER_DB_PASSWORD", "postgres"),
		DbName:        GetEnv("USER_DB_NAME", "user_service"),
		DbPort:        GetEnv("USER_DB_PORT", "5432"),
		CacheHost:     GetEnv("USER_CACHE_HOST", "localhost"),
		CachePort:     GetEnv("USER_CACHE_PORT", "6379"),
		CachePassword: GetEnv("USER_CACHE_PASSWORD", ""),
		CacheDbToken:  GetEnv("USER_CACHE_DB_TOKEN", "0"),
		GrpcAddr:      GetEnv("USER_GRPC_ADDR", ":50051"),
		HttpAddr:      GetEnv("USER_HTTP_ADDR", ":8081"),
		JwtSecret:     GetEnv("JWT_SECRET", "your-super-secret-key-change-this-in-production"),
	}
}

// LoadBookServiceConfig loads configuration for the book service
func LoadBookServiceConfig() ServiceConfig {
	return ServiceConfig{
		DbHost:     GetEnv("BOOK_DB_HOST", "localhost"),
		DbUser:     GetEnv("BOOK_DB_USER", "postgres"),
		DbPassword: GetEnv("BOOK_DB_PASSWORD", "postgres"),
		DbName:     GetEnv("BOOK_DB_NAME", "book_service"),
		DbPort:     GetEnv("BOOK_DB_PORT", "5432"),
		GrpcAddr:   GetEnv("BOOK_GRPC_ADDR", ":50052"),
		HttpAddr:   GetEnv("BOOK_HTTP_ADDR", ":8082"),
	}
}

// LoadTransactionServiceConfig loads configuration for the transaction service
func LoadTransactionServiceConfig() ServiceConfig {
	return ServiceConfig{
		DbHost:     GetEnv("TRANSACTION_DB_HOST", "localhost"),
		DbUser:     GetEnv("TRANSACTION_DB_USER", "postgres"),
		DbPassword: GetEnv("TRANSACTION_DB_PASSWORD", "postgres"),
		DbName:     GetEnv("TRANSACTION_DB_NAME", "transaction_service"),
		DbPort:     GetEnv("TRANSACTION_DB_PORT", "5432"),
		GrpcAddr:   GetEnv("TRANSACTION_GRPC_ADDR", ":50053"),
		HttpAddr:   GetEnv("TRANSACTION_HTTP_ADDR", ":8083"),
	}
}
