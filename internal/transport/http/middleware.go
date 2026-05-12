package http

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go-usersvc-demo/internal/auth"

	"github.com/gin-gonic/gin"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format(time.RFC822),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

func NewAuthMiddleware(tokenManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[AUTH] Missing Authorization header")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized - missing token"})
			return
		}

		token := authHeader

		token = strings.TrimPrefix(authHeader, "Bearer ")

		if token == "" {
			log.Printf("[AUTH] Empty token")
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized - empty token"})
			return
		}

		log.Printf("[AUTH] Validating token: %s", token[:min(len(token), 30)]+"...")

		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			log.Printf("[AUTH] Token validation failed: %v", err)
			c.AbortWithStatusJSON(401, gin.H{"error": fmt.Sprintf("Unauthorized - %s", err.Error())})
			return
		}

		log.Printf("[AUTH] Token validated for userID: %d", userID)
		ctx := ContextWithUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
