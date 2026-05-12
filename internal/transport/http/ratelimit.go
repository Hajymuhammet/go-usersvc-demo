package http

import (
	"go-usersvc-demo/internal/pkg/logger"
	"go-usersvc-demo/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)


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

func AuthenticatedRateLimitMiddleware(limiter *ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetString("request_id")
		var key string

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
