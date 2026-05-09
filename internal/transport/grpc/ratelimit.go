package grpc

import (
	"context"
	"go-usersvc-demo/internal/pkg/logger"
	"go-usersvc-demo/internal/pkg/ratelimit"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryRateLimitInterceptor creates a unary interceptor for gRPC rate limiting.
// It uses the client IP address as the key for rate limiting.
func UnaryRateLimitInterceptor(limiter *ratelimit.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get client IP from peer
		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		} else {
			clientIP = "unknown"
		}

		if !limiter.Allow(clientIP) {
			logger.Get().Warn("gRPC rate limit exceeded",
				"client_ip", clientIP,
				"method", info.FullMethod,
			)

			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
		}

		return handler(ctx, req)
	}
}

// StreamRateLimitInterceptor creates a stream interceptor for gRPC rate limiting.
// It uses the client IP address as the key for rate limiting.
func StreamRateLimitInterceptor(limiter *ratelimit.Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Get client IP from peer
		var clientIP string
		if p, ok := peer.FromContext(ss.Context()); ok {
			clientIP = p.Addr.String()
		} else {
			clientIP = "unknown"
		}

		if !limiter.Allow(clientIP) {
			logger.Get().Warn("gRPC rate limit exceeded",
				"client_ip", clientIP,
				"method", info.FullMethod,
			)

			return status.Error(codes.ResourceExhausted, "rate limit exceeded, please try again later")
		}

		return handler(srv, ss)
	}
}
