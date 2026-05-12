package service

import (
	"context"
	"errors"
	"fmt"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/pkg/logger"

	"golang.org/x/crypto/bcrypt"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	repo         domain.UserRepository
	cache        domain.UserCache
	emailService domain.EmailService
}

func NewUserService(repo domain.UserRepository, cache domain.UserCache) *UserService {
	return &UserService{repo: repo, cache: cache}
}

func (s *UserService) SetEmailService(emailService domain.EmailService) {
	s.emailService = emailService
}

func (s *UserService) CreateUser(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.NewConflictError("email already registered", input.Email)
	}
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

	_ = s.cache.Set(ctx, created)

	if s.emailService != nil {
		_ = s.emailService.SendWelcomeEmail(ctx, created.Email, created.Name)
	}

	return created, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	cached, err := s.cache.Get(ctx, id)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, redis.Nil) {
		logger.Get().Debug("cache miss", "id", id, "error", err)
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewNotFoundError("user not found", formatInt(id))
	}

	_ = s.cache.Set(ctx, user)

	return user, nil
}

func (s *UserService) ListUsers(ctx context.Context, filter domain.ListFilter) (*domain.UserList, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	return s.repo.List(ctx, filter)
}

func (s *UserService) UpdateUser(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
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

	_ = s.cache.Delete(ctx, id)
	_ = s.cache.Set(ctx, updated)

	return updated, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return domain.NewNotFoundError("user not found", formatInt(id))
	}
	_ = s.cache.Delete(ctx, id)
	return nil
}

func formatInt(i int64) string {
	return fmt.Sprintf("%d", i)
}
