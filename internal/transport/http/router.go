package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "go-usersvc-demo/docs"
	"go-usersvc-demo/internal/pkg/ratelimit"
)

func NewRouter(h *Handler, authMiddleware gin.HandlerFunc, rateLimiter *ratelimit.Limiter, authRateLimiter *ratelimit.Limiter) *gin.Engine {
	r := gin.New()

	r.Use(RequestIDMiddleware())
	r.Use(RequestLoggingMiddleware())
	r.Use(CORSMiddleware())
	r.Use(RecoveryMiddleware())
	r.Use(RateLimitMiddleware(rateLimiter))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/users", h.CreateUser)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)

	users := r.Group("/users")
	users.Use(authMiddleware)
	users.Use(AuthenticatedRateLimitMiddleware(authRateLimiter))
	{
		users.GET("/me", h.GetProfile)
		users.GET("", h.ListUsers)
		users.GET("/:id", h.GetUserByID)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	email := r.Group("/email")
	email.Use(authMiddleware)
	email.Use(AuthenticatedRateLimitMiddleware(authRateLimiter))
	{
		email.POST("/send", h.SendEmail)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
