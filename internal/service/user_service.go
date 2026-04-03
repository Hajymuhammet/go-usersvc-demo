package service

import (
	"context"
	"errors"
	"fmt"

	"go-usersvc-demo/internal/domain"

	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"
)

// UserService handles business logic for users.
type UserService struct {
	repo  domain.UserRepository
	cache domain.UserCache
}

// NewUserService creates a new UserService.
func NewUserService(repo domain.UserRepository, cache domain.UserCache) *UserService {
	return &UserService{repo: repo, cache: cache}
}

// CreateUser hashes the password and persists the new user.
func (s *UserService) CreateUser(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	// Check for duplicate email
	existing, _ := s.repo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}

	user := &domain.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Warm the cache
	_ = s.cache.Set(ctx, created)

	return created, nil
}

// GetUserByID returns a user, preferring the cache over the database.
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	// Try cache first
	cached, err := s.cache.Get(ctx, id)
	if err == nil {
		return cached, nil
	}

	// Cache miss — fall through to DB
	if !errors.Is(err, redis.Nil) {
		// Log non-Nil cache errors but continue
		_ = err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Hydrate cache for next request
	_ = s.cache.Set(ctx, user)

	return user, nil
}

// ListUsers returns a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, filter domain.ListFilter) (*domain.UserList, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.repo.List(ctx, filter)
}

// UpdateUser updates allowed user fields and invalidates the cache.
func (s *UserService) UpdateUser(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	// Hash password if provided
	if input.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("service: hash password: %w", err)
		}
		h := string(hashed)
		input.Password = &h
	}

	updated, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	// Invalidate stale cache entry
	_ = s.cache.Delete(ctx, id)
	_ = s.cache.Set(ctx, updated)

	return updated, nil
}

// DeleteUser removes a user and evicts the cache.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, id)
	return nil
}
