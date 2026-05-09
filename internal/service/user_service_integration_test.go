package service_test

import (
	"context"
	"testing"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/service"
	"go-usersvc-demo/internal/testhelpers"
)

func TestUserService_CreateUser_WithMocks(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()
	input := domain.CreateUserInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	user, err := userService.CreateUser(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Name != input.Name {
		t.Errorf("Expected name %s, got %s", input.Name, user.Name)
	}
	if user.Email != input.Email {
		t.Errorf("Expected email %s, got %s", input.Email, user.Email)
	}
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()
	input := domain.CreateUserInput{
		Name:     "User 1",
		Email:    "duplicate@example.com",
		Password: "pass123",
	}

	// Create first user
	_, err := userService.CreateUser(ctx, input)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	// Try to create with same email
	input.Name = "User 2"
	_, err = userService.CreateUser(ctx, input)
	if err == nil {
		t.Fatal("Expected error for duplicate email")
	}
}

func TestUserService_GetUser_FromCache(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()
	input := domain.CreateUserInput{
		Name:     "Cache Test User",
		Email:    "cache@example.com",
		Password: "pass123",
	}

	// Create user
	createdUser, _ := userService.CreateUser(ctx, input)

	// Set in cache
	mockCache.Set(ctx, createdUser)

	// Get user - should fetch from cache
	retrieved, err := userService.GetUserByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.ID != createdUser.ID {
		t.Errorf("Expected user ID %d, got %d", createdUser.ID, retrieved.ID)
	}
}

func TestUserService_ListUsers_WithMocks(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()

	// Create multiple users
	for i := 0; i < 3; i++ {
		input := domain.CreateUserInput{
			Name:     "User " + string(rune(i)),
			Email:    "user" + string(rune(i)) + "@example.com",
			Password: "pass123",
		}
		userService.CreateUser(ctx, input)
	}

	// List users
	list, err := userService.ListUsers(ctx, domain.ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(list.Data) != 3 {
		t.Errorf("Expected 3 users, got %d", len(list.Data))
	}
}

func TestUserService_UpdateUser_WithMocks(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()
	input := domain.CreateUserInput{
		Name:     "Original Name",
		Email:    "update@example.com",
		Password: "pass123",
	}

	user, _ := userService.CreateUser(ctx, input)

	// Update user
	newName := "Updated Name"
	updated, err := userService.UpdateUser(ctx, user.ID, domain.UpdateUserInput{
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if updated.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updated.Name)
	}
}

func TestUserService_DeleteUser_WithMocks(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()
	input := domain.CreateUserInput{
		Name:     "Delete Me",
		Email:    "delete@example.com",
		Password: "pass123",
	}

	user, _ := userService.CreateUser(ctx, input)

	// Delete user
	err := userService.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify deleted
	_, err = userService.GetUserByID(ctx, user.ID)
	if err == nil {
		t.Error("Expected error after deletion")
	}
}

func TestUserService_RepositoryError_Handling(t *testing.T) {
	mockRepo := testhelpers.NewMockUserRepository()
	mockCache := testhelpers.NewMockUserCache()
	userService := service.NewUserService(mockRepo, mockCache)

	ctx := context.Background()

	// Inject error
	testErr := domain.NewNotFoundError("user not found", "user with id 999 not found")
	mockRepo.SetupError(testErr)
	defer mockRepo.ClearError()

	_, err := userService.GetUserByID(ctx, 999)
	if err == nil {
		t.Error("Expected error from injected failure")
	}
}
