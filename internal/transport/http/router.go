package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	"github.com/swaggo/gin-swagger"        // gin-swagger middleware

)

// NewRouter creates and returns a configured Gin engine.
func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	users := r.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.ListUsers)
		users.GET("/:id", h.GetUserByID)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
