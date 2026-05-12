package grpc

import (
	"context"
	"log"
	"strings"
	"time"

	"go-usersvc-demo/internal/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	log.Printf("[%s] %s %v %s", time.Now().Format(time.RFC822), info.FullMethod, duration, status.Code(err).String())
	return resp, err
}

func NewAuthInterceptor(tokenManager *auth.Manager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		tokens := md["authorization"]
		if len(tokens) == 0 || tokens[0] == "" {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		token := tokens[0]

		token = strings.TrimPrefix(token, "Bearer ")

		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userIDContextKey, userID)
		return handler(ctx, req)
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
