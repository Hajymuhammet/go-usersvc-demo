package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	AppEnv   string
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

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	appEnv := getEnv("APP_ENV", "development")
	authSecret := getEnv("AUTH_SECRET", "")

	if authSecret == "" {
		return nil, fmt.Errorf("AUTH_SECRET environment variable must be set and non-empty")
	}

	if authSecret == "change-this-secret" {
		return nil, fmt.Errorf("AUTH_SECRET must be changed from the default value. Generate a strong secret using: openssl rand -base64 32")
	}

	cfg := &Config{
		AppEnv: appEnv,
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
			Secret:          authSecret,
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
