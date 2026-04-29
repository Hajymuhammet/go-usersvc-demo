package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string             `json:"status"`
	Timestamp time.Time          `json:"timestamp"`
	Services  map[string]Service `json:"services"`
}

// Service represents a service health status.
type Service struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthCheck checks the health of all services.
func HealthCheck(db *pgxpool.Pool, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		response := HealthResponse{
			Timestamp: time.Now(),
			Services:  make(map[string]Service),
		}

		// Check Database
		if err := db.Ping(ctx); err != nil {
			response.Services["database"] = Service{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			response.Services["database"] = Service{Status: "up"}
		}

		// Check Redis
		if err := redisClient.Ping(ctx).Err(); err != nil {
			response.Services["cache"] = Service{
				Status: "down",
				Error:  err.Error(),
			}
		} else {
			response.Services["cache"] = Service{Status: "up"}
		}

		// Determine overall status
		allUp := true
		for _, svc := range response.Services {
			if svc.Status != "up" {
				allUp = false
				break
			}
		}

		if allUp {
			response.Status = "healthy"
			c.JSON(http.StatusOK, response)
		} else {
			response.Status = "degraded"
			c.JSON(http.StatusServiceUnavailable, response)
		}
	}
}
