package service

import (
	"context"
	"errors"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/pkg/logger"

	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"
)

// UserService handles business logic for users.
type UserService struct {
	repo         domain.UserRepository
	cache        domain.UserCache
	emailService *EmailService
}

// NewUserService creates a new UserService.
func NewUserService(repo domain.UserRepository, cache domain.UserCache) *UserService {
	return &UserService{repo: repo, cache: cache}
}

// SetEmailService sets the email service for UserService.
func (s *UserService) SetEmailService(emailService *EmailService) {
	s.emailService = emailService
}

// CreateUser hashes the password and persists the new user.
func (s *UserService) CreateUser(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	// Check for duplicate email
	existing, _ := s.repo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.NewConflictError("email already registered", input.Email)
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain.NewInternalError("failed to hash password", err)
	}

	user := &domain.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashed),
	}

	created, err := s.repo.Create(ctx, user)
	if err != nil {
		logger.Get().Error("failed to create user", "error", err)
		return nil, domain.NewInternalError("failed to create user", err)
	}

	// Warm the cache
	_ = s.cache.Set(ctx, created)

	// Send welcome email (non-blocking, log errors but don't fail the operation)
	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendWelcomeEmail(context.Background(), created.Email, created.Name); err != nil {
				logger.Get().Error("failed to send welcome email", "error", err, "email", created.Email)
			}
		}()
	}

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
		logger.Get().Debug("cache miss", "id", id, "error", err)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewNotFoundError("user not found", formatInt(id))
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
			return nil, domain.NewInternalError("failed to hash password", err)
		}
		h := string(hashed)
		input.Password = &h
	}

	updated, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, domain.NewNotFoundError("user not found", formatInt(id))
	}

	// Invalidate stale cache entry
	_ = s.cache.Delete(ctx, id)
	_ = s.cache.Set(ctx, updated)

	return updated, nil
}

// DeleteUser removes a user and evicts the cache.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.NewNotFoundError("user not found", formatInt(id))
	}
	_ = s.cache.Delete(ctx, id)
	return nil
}

// formatInt is a helper to convert int64 to string for error details.
func formatInt(i int64) string {
	return ""
}
