package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-usersvc-demo/internal/auth"
	"go-usersvc-demo/internal/config"
	"go-usersvc-demo/internal/infrastructure/postgres"
	infraredis "go-usersvc-demo/internal/infrastructure/redis"
	"go-usersvc-demo/internal/pkg/logger"
	"go-usersvc-demo/internal/pkg/ratelimit"
	"go-usersvc-demo/internal/service"
	transportgrpc "go-usersvc-demo/internal/transport/grpc"
	transporthttp "go-usersvc-demo/internal/transport/http"
	"go-usersvc-demo/pkg/pb"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// @title           Go User Service Demo API
// @version         1.0
// @description     A production-ready user management service with REST and gRPC APIs.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Initialize logger
	isDev := cfg.AppEnv == "development"
	logger.Initialize(isDev)
	log := logger.Get()

	log.Info("starting application",
		"env", cfg.AppEnv,
		"port", cfg.Server.Port,
		"grpc_port", cfg.Server.GRPCPort,
	)

	// Connect to PostgreSQL
	db, err := postgres.NewPool(cfg.Database.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to postgres")

	// Connect to Redis
	redisClient, err := infraredis.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	log.Info("connected to redis")

	// Wire dependencies
	userRepo := postgres.NewUserRepo(db)
	userCache := infraredis.NewUserCache(redisClient)
	userSvc := service.NewUserService(userRepo, userCache)

	// Initialize email service with mock provider (can be replaced with real provider)
	emailProvider := service.NewMockEmailProvider()
	emailSvc := service.NewEmailService(emailProvider)

	tokenManager := auth.NewManager(
		cfg.Auth.Secret,
		parseDurationOrDefault(cfg.Auth.AccessTokenTTL, 15*time.Minute),
		parseDurationOrDefault(cfg.Auth.RefreshTokenTTL, 168*time.Hour),
	)
	authSvc := service.NewAuthService(userRepo, tokenManager)

	// Initialize rate limiters
	// Public endpoints: 100 requests/sec per IP, burst of 10
	publicRateLimiter := ratelimit.NewLimiter(100, 10)
	// Authenticated endpoints: 50 requests/sec per user, burst of 5
	authRateLimiter := ratelimit.NewLimiter(50, 5)
	log.Info("rate limiters initialized")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	// 1. Start REST Server
	httpHandler := transporthttp.NewHandler(userSvc, authSvc, emailSvc)
	router := transporthttp.NewRouter(
		httpHandler,
		transporthttp.NewAuthMiddleware(tokenManager),
		publicRateLimiter,
		authRateLimiter,
	)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
		Handler: router,
	}

	g.Go(func() error {
		log.Info("rest server started", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	// 2. Start gRPC Server
	grpcRateLimiter := ratelimit.NewLimiter(100, 10)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			transportgrpc.LoggingInterceptor,
			transportgrpc.NewAuthInterceptor(tokenManager),
			transportgrpc.UnaryRateLimitInterceptor(grpcRateLimiter),
		),
		grpc.ChainStreamInterceptor(
			transportgrpc.StreamRateLimitInterceptor(grpcRateLimiter),
		),
	)
	pb.RegisterUserServiceServer(grpcServer, transportgrpc.NewHandler(userSvc))
	reflection.Register(grpcServer)

	g.Go(func() error {
		addr := fmt.Sprintf(":%s", cfg.Server.GRPCPort)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("grpc listen: %w", err)
		}
		log.Info("grpc server started", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			return fmt.Errorf("grpc serve: %w", err)
		}
		return nil
	})

	// 3. Graceful Shutdown
	g.Go(func() error {
		<-ctx.Done()
		log.Info("shutdown signal received, gracefully shutting down servers")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("http server shutdown error", "error", err)
		}
		grpcServer.GracefulStop()

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error("service error", "error", err)
		os.Exit(1)
	}
	log.Info("service stopped gracefully")
}

func parseDurationOrDefault(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}
