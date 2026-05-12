package main

import (
	"fmt"
	"log"
	"time"

	"go-usersvc-demo/internal/auth"
	"go-usersvc-demo/internal/config"
	"go-usersvc-demo/internal/infrastructure/postgres"
	infraredis "go-usersvc-demo/internal/infrastructure/redis"
	"go-usersvc-demo/internal/pkg/ratelimit"
	"go-usersvc-demo/internal/service"
	transporthttp "go-usersvc-demo/internal/transport/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Connect to PostgreSQL
	db, err := postgres.NewPool(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to PostgreSQL")

	// Connect to Redis
	redisClient, err := infraredis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✅ Connected to Redis")

	userRepo := postgres.NewUserRepo(db)
	userCache := infraredis.NewUserCache(redisClient)
	userSvc := service.NewUserService(userRepo, userCache)

	emailProvider := service.NewMockEmailProvider()
	emailSvc := service.NewEmailService(emailProvider)

	tokenManager := auth.NewManager(
		cfg.Auth.Secret,
		parseDurationOrDefault(cfg.Auth.AccessTokenTTL, 15*time.Minute),
		parseDurationOrDefault(cfg.Auth.RefreshTokenTTL, 168*time.Hour),
	)
	authSvc := service.NewAuthService(userRepo, tokenManager)

	publicRateLimiter := ratelimit.NewLimiter(100, 10)
	authRateLimiter := ratelimit.NewLimiter(50, 5)
	log.Println("✅ Rate limiters initialized")
	handler := transporthttp.NewHandler(userSvc, authSvc, emailSvc)
	router := transporthttp.NewRouter(
		handler,
		transporthttp.NewAuthMiddleware(tokenManager),
		publicRateLimiter,
		authRateLimiter,
	)

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("🚀 REST server listening on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func parseDurationOrDefault(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
