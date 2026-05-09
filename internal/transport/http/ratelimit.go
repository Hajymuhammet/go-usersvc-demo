package http

import (
	"go-usersvc-demo/internal/pkg/logger"
	"go-usersvc-demo/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates a rate limit middleware for HTTP handlers.
// It uses the client IP address as the key for rate limiting.
func RateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		requestID := c.GetString("request_id")

		if !limiter.Allow(clientIP) {
			logger.Get().Warn("rate limit exceeded",
				"request_id", requestID,
				"client_ip", clientIP,
				"path", c.Request.URL.Path,
			)

			c.JSON(429, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthenticatedRateLimitMiddleware creates a rate limit middleware for authenticated endpoints.
// It uses the user ID as the key for rate limiting when available, otherwise falls back to IP.
func AuthenticatedRateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetString("request_id")
		var key string

		// Use user ID if available, otherwise use client IP
		if userID, ok := UserIDFromContext(c.Request.Context()); ok {
			key = "user:" + string(rune(userID))
		} else {
			key = "ip:" + c.ClientIP()
		}

		if !limiter.Allow(key) {
			logger.Get().Warn("rate limit exceeded",
				"request_id", requestID,
				"key", key,
				"path", c.Request.URL.Path,
			)

			c.JSON(429, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "too many requests, please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
