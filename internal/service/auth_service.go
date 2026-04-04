package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"go-usersvc-demo/internal/auth"
	"go-usersvc-demo/internal/domain"
)

type AuthService struct {
	repo         domain.UserRepository
	tokenManager *auth.Manager
	rateLimiter  *RateLimiter
	logger       *log.Logger
}

type RateLimiter struct {
	attempts map[string]*loginAttempt
	mu       sync.RWMutex
	maxAttempts int
	window     time.Duration
}

type loginAttempt struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts:   make(map[string]*loginAttempt),
		maxAttempts: maxAttempts,
		window:     window,
	}
}

func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	attempt, exists := rl.attempts[key]

	if !exists || now.After(attempt.resetTime) {
		rl.attempts[key] = &loginAttempt{count: 1, resetTime: now.Add(rl.window)}
		return true
	}

	if attempt.count >= rl.maxAttempts {
		return false
	}

	attempt.count++
	return true
}

func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

func NewAuthService(repo domain.UserRepository, tokenManager *auth.Manager) *AuthService {
	return &AuthService{
		repo:        repo,
		tokenManager: tokenManager,
		rateLimiter: NewRateLimiter(5, 15*time.Minute), // 5 attempts per 15 minutes
		logger:      log.New(log.Writer(), "[AUTH] ", log.LstdFlags),
	}
}

// ValidatePassword checks if password meets security requirements
func (s *AuthService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return errors.New("password must be less than 128 characters")
	}

	// Check for at least one uppercase, one lowercase, one digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	// Check for common weak passwords (basic check)
	weakPasswords := []string{"password", "123456", "qwerty", "admin", "letmein"}
	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			return errors.New("password is too common, please choose a stronger password")
		}
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, *domain.User, error) {
	// Rate limiting check
	if !s.rateLimiter.IsAllowed(email) {
		s.logger.Printf("Rate limit exceeded for email: %s", email)
		return "", "", nil, errors.New("too many login attempts, please try again later")
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err.Error() == "user not found" {
			s.logger.Printf("Login failed: user not found for email: %s", email)
			return "", "", nil, errors.New("invalid credentials")
		}
		return "", "", nil, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.Printf("Login failed: invalid password for user ID: %d", user.ID)
		return "", "", nil, errors.New("invalid credentials")
	}

	// Reset rate limiter on successful login
	s.rateLimiter.Reset(email)
	s.logger.Printf("Login successful for user ID: %d", user.ID)

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("login: generate access token: %w", err)
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, fmt.Errorf("login: generate refresh token: %w", err)
	}

	return accessToken, refreshToken, user, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := s.tokenManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		s.logger.Printf("Refresh failed: invalid token")
		return "", "", errors.New("invalid refresh token")
	}

	s.logger.Printf("Token refresh for user ID: %d", userID)

	accessToken, err := s.tokenManager.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("refresh: generate access token: %w", err)
	}

	newRefreshToken, err := s.tokenManager.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("refresh: generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// ChangePassword allows authenticated users to change their password
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Printf("Change password failed: user not found for ID: %d", userID)
		return fmt.Errorf("change password: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		s.logger.Printf("Change password failed: invalid old password for user ID: %d", userID)
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		s.logger.Printf("Change password failed: invalid new password for user ID: %d", userID)
		return fmt.Errorf("new password validation: %w", err)
	}

	// Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("change password: hash new password: %w", err)
	}

	// Update user password
	passwordStr := string(hashed)
	updateInput := domain.UpdateUserInput{
		Password: &passwordStr,
	}
	_, err = s.repo.Update(ctx, userID, updateInput)
	if err != nil {
		return fmt.Errorf("change password: update user: %w", err)
	}

	s.logger.Printf("Password changed successfully for user ID: %d", userID)
	return nil
}

// Logout could invalidate tokens (requires token storage for blacklisting)
func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	// In a real implementation, you would:
	// 1. Add tokens to a blacklist/revoked set in Redis
	// 2. Check tokens against blacklist during validation
	// For now, this is a placeholder
	return nil
}
