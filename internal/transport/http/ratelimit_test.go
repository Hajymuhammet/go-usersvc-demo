package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-usersvc-demo/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddleware_Allow(t *testing.T) {
	limiter := ratelimit.NewLimiter(100, 10)
	defer limiter.ResetAll()

	middleware := RateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRateLimitMiddleware_Denied(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	middleware := RateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	clientIP := "192.168.1.100"

	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = clientIP + ":8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != 200 {
		t.Errorf("Expected first request status 200, got %d", w1.Code)
	}

	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = clientIP + ":8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Expected status 429 (rate limited), got %d", w2.Code)
	}
}

func TestRateLimitMiddleware_Different_IPs(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	middleware := RateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w1.Code != 200 {
		t.Errorf("Expected IP1 request to succeed, got status %d", w1.Code)
	}
	if w2.Code != 200 {
		t.Errorf("Expected IP2 request to succeed, got status %d", w2.Code)
	}
}

func TestAuthenticatedRateLimitMiddleware_WithUserID(t *testing.T) {
	limiter := ratelimit.NewLimiter(100, 10)
	defer limiter.ResetAll()

	middleware := AuthenticatedRateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(ContextWithUserID(c.Request.Context(), 123))
		c.Next()
	})
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200 for authenticated request, got %d", w.Code)
	}
}

func TestAuthenticatedRateLimitMiddleware_RateLimitExceeded(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 1)
	defer limiter.ResetAll()

	middleware := AuthenticatedRateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(ContextWithUserID(c.Request.Context(), 123))
		c.Next()
	})
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "127.0.0.1:8080"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != 200 {
		t.Errorf("Expected first request to succeed, got status %d", w1.Code)
	}

	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "127.0.0.1:8080"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Errorf("Expected second request to be rate limited (429), got status %d", w2.Code)
	}
}

func TestRateLimitMiddleware_Response_Format(t *testing.T) {
	limiter := ratelimit.NewLimiter(1, 0)
	defer limiter.ResetAll()

	middleware := RateLimitMiddleware(limiter)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	clientIP := "127.0.0.1:9000"

	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = clientIP
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = clientIP
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != 429 {
		t.Fatalf("Expected rate limited status, got %d", w2.Code)
	}

	if w2.Body.String() == "" {
		t.Error("Expected non-empty response body")
	}
}
