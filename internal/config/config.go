package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port     string
	GRPCPort string
}

type AuthConfig struct {
	Secret          string
	AccessTokenTTL  string
	RefreshTokenTTL string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Load reads configuration from environment (optionally from .env file).
func Load() (*Config, error) {
	// Load .env if it exists (ignore error if not present)
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	cfg := &Config{
		Server: ServerConfig{
			Port:     getEnv("SERVER_PORT", "8080"),
			GRPCPort: getEnv("GRPC_PORT", "50051"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "secret"),
			Name:     getEnv("DB_NAME", "usersdb"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Auth: AuthConfig{
			Secret:          getEnv("AUTH_SECRET", "change-this-secret"),
			AccessTokenTTL:  getEnv("AUTH_ACCESS_TOKEN_TTL", "15m"),
			RefreshTokenTTL: getEnv("AUTH_REFRESH_TOKEN_TTL", "168h"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
