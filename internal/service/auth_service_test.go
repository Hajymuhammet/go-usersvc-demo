package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-usersvc-demo/internal/auth"
	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/domain/mocks"
	"go-usersvc-demo/internal/service"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_ValidatePassword_Success(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("ValidPass123")
	require.NoError(t, err)
}

func TestAuthService_ValidatePassword_TooShort(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("Short1")
	require.Error(t, err)
}

func TestAuthService_ValidatePassword_TooLong(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	longPassword := "A1bcdefghijklmnopqrstuvwxyz"
	for len(longPassword) < 130 {
		longPassword += "A1bcdefghijklmnopqrstuvwxyz"
	}

	err := authService.ValidatePassword(longPassword)
	require.Error(t, err)
}

func TestAuthService_ValidatePassword_NoUppercase(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("lowercase123")
	require.Error(t, err)
}

func TestAuthService_ValidatePassword_NoLowercase(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("UPPERCASE123")
	require.Error(t, err)
}

func TestAuthService_ValidatePassword_NoDigit(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("NoDigitHere")
	require.Error(t, err)
}

func TestAuthService_ValidatePassword_WeakPassword(t *testing.T) {
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	repo := mocks.NewUserRepository(t)
	authService := service.NewAuthService(repo, tokenManager)

	err := authService.ValidatePassword("Password123")
	require.Error(t, err)
}

func TestRateLimiter_AllowedAttempts(t *testing.T) {
	rl := service.NewRateLimiter(3, 1*time.Minute)

	for i := 0; i < 3; i++ {
		require.True(t, rl.IsAllowed("user1"))
	}
	require.False(t, rl.IsAllowed("user1"))
}

func TestRateLimiter_MultipleUsers(t *testing.T) {
	rl := service.NewRateLimiter(2, 1*time.Minute)

	for i := 0; i < 2; i++ {
		require.True(t, rl.IsAllowed("user1"))
	}
	require.False(t, rl.IsAllowed("user1"))
	require.True(t, rl.IsAllowed("user2"))
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := service.NewRateLimiter(2, 1*time.Minute)

	for i := 0; i < 2; i++ {
		require.True(t, rl.IsAllowed("user1"))
	}
	require.False(t, rl.IsAllowed("user1"))

	rl.Reset("user1")
	require.True(t, rl.IsAllowed("user1"))
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	password := "StrongPass123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	repo.EXPECT().GetByEmail(mock.Anything, user.Email).Return(user, nil)

	accessToken, refreshToken, loginUser, err := authService.Login(context.Background(), user.Email, password)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
	require.Equal(t, user.ID, loginUser.ID)

	validatedID, err := tokenManager.ValidateAccessToken(accessToken)
	require.NoError(t, err)
	require.Equal(t, user.ID, validatedID)
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	repo.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errors.New("user not found"))

	_, _, _, err := authService.Login(context.Background(), "nonexistent@example.com", "SomePass123")
	require.Error(t, err)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	password := "StrongPass123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &domain.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	repo.EXPECT().GetByEmail(mock.Anything, user.Email).Return(user, nil)

	_, _, _, err := authService.Login(context.Background(), user.Email, "WrongPass123")
	require.Error(t, err)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	userID := int64(456)
	refreshToken, _ := tokenManager.GenerateRefreshToken(userID)

	newAccessToken, newRefreshToken, err := authService.Refresh(context.Background(), refreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, newAccessToken)
	require.NotEmpty(t, newRefreshToken)

	validatedID, err := tokenManager.ValidateAccessToken(newAccessToken)
	require.NoError(t, err)
	require.Equal(t, userID, validatedID)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	repo := mocks.NewUserRepository(t)
	tokenManager := auth.NewManager("test-secret", 1*time.Hour, 7*24*time.Hour)
	authService := service.NewAuthService(repo, tokenManager)

	_, _, err := authService.Refresh(context.Background(), "invalid-token")
	require.Error(t, err)
}
