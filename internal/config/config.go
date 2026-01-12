package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Storage  StorageConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type StorageConfig struct {
	BasePath    string
	MaxFileSize int64
	BaseURL     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		refreshExpiry = 168 * time.Hour
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	maxFileSize, err := strconv.ParseInt(getEnv("STORAGE_MAX_FILE_SIZE", "104857600"), 10, 64)
	if err != nil {
		maxFileSize = 100 * 1024 * 1024 // 100MB
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "messenger"),
			Password: getEnv("DB_PASS", "password"),
			Name:     getEnv("DB_NAME", "messenger_db"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Storage: StorageConfig{
			BasePath:    getEnv("STORAGE_BASE_PATH", "./uploads"),
			MaxFileSize: maxFileSize,
			BaseURL:     getEnv("STORAGE_BASE_URL", "/files"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "default-secret-change-me"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
