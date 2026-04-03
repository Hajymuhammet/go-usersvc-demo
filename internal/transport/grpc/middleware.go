package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC requests.
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	log.Printf("[%s] %s %v %s", time.Now().Format(time.RFC822), info.FullMethod, duration, status.Code(err).String())
	return resp, err
}

// AuthInterceptor is a placeholder for authentication.
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Example: Check metadata for token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	token := md["authorization"]
	if len(token) == 0 || token[0] == "" {
		return nil, status.Error(codes.Unauthenticated, "missing token")
	}

	// Validate token here
	// For demo, just pass
	return handler(ctx, req)
}
