package http

import (
	"context"
	"fmt"
	"time"

	"go-usersvc-demo/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RequestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := c.GetString("request_id")
		logger.Get().Info("incoming request",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_addr", c.ClientIP(),
		)

		c.Next()

		duration := time.Since(startTime)
		logger.Get().Info("request completed",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		requestID := c.GetString("request_id")
		logger.Get().Error("panic recovered",
			"request_id", requestID,
			"panic", fmt.Sprintf("%v", err),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		c.JSON(500, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "an unexpected error occurred",
		})
	})
}
