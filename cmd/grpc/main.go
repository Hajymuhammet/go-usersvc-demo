package main

import (
	"fmt"
	"log"
	"net"

	"go-usersvc-demo/internal/config"
	"go-usersvc-demo/internal/infrastructure/postgres"
	infraredis "go-usersvc-demo/internal/infrastructure/redis"
	"go-usersvc-demo/internal/service"
	transportgrpc "go-usersvc-demo/internal/transport/grpc"
	"go-usersvc-demo/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Wire dependencies
	userRepo := postgres.NewUserRepo(db)
	userCache := infraredis.NewUserCache(redisClient)
	userSvc := service.NewUserService(userRepo, userCache)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(transportgrpc.LoggingInterceptor),
	)
	pb.RegisterUserServiceServer(grpcServer, transportgrpc.NewHandler(userSvc))
	reflection.Register(grpcServer) // enables grpcurl discovery

	addr := fmt.Sprintf(":%s", cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Printf("⚡ gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc server: %v", err)
	}
}
