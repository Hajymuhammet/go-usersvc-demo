package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware

	_ "go-usersvc-demo/docs" // swagger generated docs
)

// NewRouter creates and returns a configured Gin engine.
func NewRouter(h *Handler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.New() // Use gin.New() instead of gin.Default() to avoid default middlewares

	// Add custom middlewares
	r.Use(LoggingMiddleware())
	r.Use(gin.Recovery()) // Keep recovery

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public endpoints
	r.POST("/users", h.CreateUser)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/refresh", h.Refresh)

	// Authenticated user routes
	users := r.Group("/users")
	users.Use(authMiddleware)
	{
		users.GET("/me", h.GetProfile)
		users.GET("", h.ListUsers)
		users.GET("/:id", h.GetUserByID)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
