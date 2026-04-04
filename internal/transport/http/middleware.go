package http

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-usersvc-demo/internal/auth"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// LoggingMiddleware logs HTTP requests.
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

// NewAuthMiddleware validates an access token and stores the authenticated user ID in request context.
func NewAuthMiddleware(tokenManager *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := tokenManager.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		ctx := ContextWithUserID(c.Request.Context(), userID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user ID stored in context.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

// ContextWithUserID stores the authenticated user ID in the provided context.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
