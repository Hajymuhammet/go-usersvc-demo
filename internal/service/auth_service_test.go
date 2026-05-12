package service_test

import (
	"context"
	"testing"
	"time"

	"go-usersvc-demo/internal/auth"
	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/service"
)

func TestAuthService_ValidatePassword_Success(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("ValidPass123")
	if err != nil {
		t.Errorf("expected valid password to pass validation, got error: %v", err)
	}
}

func TestAuthService_ValidatePassword_TooShort(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("Short1")
	if err == nil {
		t.Fatal("expected error for short password, got nil")
	}
}

func TestAuthService_ValidatePassword_TooLong(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	longPassword := "A1bcdefghijklmnopqrstuvwxyz"
	for len(longPassword) < 130 {
		longPassword += "A1bcdefghijklmnopqrstuvwxyz"
	}

	err := authService.ValidatePassword(longPassword)
	if err == nil {
		t.Fatal("expected error for long password, got nil")
	}
}

func TestAuthService_ValidatePassword_NoUppercase(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("lowercase123")
	if err == nil {
		t.Fatal("expected error for password without uppercase, got nil")
	}
}

func TestAuthService_ValidatePassword_NoLowercase(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("UPPERCASE123")
	if err == nil {
		t.Fatal("expected error for password without lowercase, got nil")
	}
}

func TestAuthService_ValidatePassword_NoDigit(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("NoDigitHere")
	if err == nil {
		t.Fatal("expected error for password without digit, got nil")
	}
}

func TestAuthService_ValidatePassword_WeakPassword(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(newMockRepo(), tokenManager)

	err := authService.ValidatePassword("Password123")
	if err == nil {
		t.Fatal("expected error for common weak password, got nil")
	}
}

func TestRateLimiter_AllowedAttempts(t *testing.T) {
	rl := service.NewRateLimiter(3, 1*time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.IsAllowed("user1") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	if rl.IsAllowed("user1") {
		t.Fatal("4th attempt should not be allowed")
	}
}

func TestRateLimiter_MultipleUsers(t *testing.T) {
	rl := service.NewRateLimiter(2, 1*time.Minute)

	for i := 0; i < 2; i++ {
		if !rl.IsAllowed("user1") {
			t.Fatalf("user1 attempt %d should be allowed", i+1)
		}
	}

	if rl.IsAllowed("user1") {
		t.Fatal("user1's 3rd attempt should not be allowed")
	}

	if !rl.IsAllowed("user2") {
		t.Fatal("user2's 1st attempt should be allowed")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := service.NewRateLimiter(2, 1*time.Minute)

	for i := 0; i < 2; i++ {
		if !rl.IsAllowed("user1") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}

	if rl.IsAllowed("user1") {
		t.Fatal("3rd attempt should not be allowed")
	}

	rl.Reset("user1")

	if !rl.IsAllowed("user1") {
		t.Fatal("attempt after reset should be allowed")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	userSvc := service.NewUserService(repo, newMockCache())
	user, err := userSvc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "StrongPass123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	accessToken, refreshToken, loginUser, err := authService.Login(context.Background(), "test@example.com", "StrongPass123")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	if loginUser.ID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, loginUser.ID)
	}

	validatedID, err := tokenManager.ValidateAccessToken(accessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if validatedID != user.ID {
		t.Errorf("expected userID %d in token, got %d", user.ID, validatedID)
	}
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	_, _, _, err := authService.Login(context.Background(), "nonexistent@example.com", "SomePass123")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	userSvc := service.NewUserService(repo, newMockCache())
	_, err := userSvc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "StrongPass123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	_, _, _, err = authService.Login(context.Background(), "test@example.com", "WrongPass123")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestAuthService_Login_RateLimiting(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	userSvc := service.NewUserService(repo, newMockCache())
	_, err := userSvc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "StrongPass123",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	email := "test@example.com"

	for i := 0; i < 5; i++ {
		_, _, _, err = authService.Login(context.Background(), email, "WrongPassword")
		if err == nil {
			t.Fatalf("login attempt %d should fail with wrong password", i+1)
		}
	}

	_, _, _, err = authService.Login(context.Background(), email, "StrongPass123")
	if err == nil || err.Error() != "too many login attempts, please try again later" {
		t.Errorf("expected rate_limited error after 5 attempts, got: %v", err)
	}
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	userID := int64(456)
	refreshToken, err := tokenManager.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	newAccessToken, newRefreshToken, err := authService.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}

	if newAccessToken == "" || newRefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	validatedID, err := tokenManager.ValidateAccessToken(newAccessToken)
	if err != nil {
		t.Fatalf("failed to validate new access token: %v", err)
	}

	if validatedID != userID {
		t.Errorf("expected userID %d in new token, got %d", userID, validatedID)
	}
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	repo := newMockRepo()
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	_, _, err := authService.Refresh(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}
