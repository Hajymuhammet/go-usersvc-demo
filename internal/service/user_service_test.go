package service_test

import (
	"context"
	"errors"
	"testing"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/domain/mocks"
	"go-usersvc-demo/internal/service"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSvc(t *testing.T) (*service.UserService, *mocks.UserRepository, *mocks.UserCache) {
	repo := mocks.NewUserRepository(t)
	cache := mocks.NewUserCache(t)
	svc := service.NewUserService(repo, cache)
	return svc, repo, cache
}

func TestCreateUser_Success(t *testing.T) {
	svc, repo, cache := newSvc(t)

	input := domain.CreateUserInput{
		Name: "John Doe", Email: "john@example.com", Password: "pass123",
	}

	repo.EXPECT().GetByEmail(mock.Anything, input.Email).Return(nil, errors.New("not found"))
	repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Name == input.Name && u.Email == input.Email
	})).Return(&domain.User{ID: 1, Name: input.Name, Email: input.Email}, nil)
	cache.EXPECT().Set(mock.Anything, mock.Anything).Return(nil)

	user, err := svc.CreateUser(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, int64(1), user.ID)
	require.Equal(t, input.Name, user.Name)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc, repo, _ := newSvc(t)

	input := domain.CreateUserInput{Name: "John", Email: "dup@example.com", Password: "pass123"}
	repo.EXPECT().GetByEmail(mock.Anything, input.Email).Return(&domain.User{ID: 1, Email: input.Email}, nil)

	_, err := svc.CreateUser(context.Background(), input)
	require.Error(t, err)
	require.Contains(t, err.Error(), "email already registered")
}

func TestGetUserByID_CacheHit(t *testing.T) {
	svc, _, cache := newSvc(t)

	expected := &domain.User{ID: 99, Name: "Cached", Email: "cache@example.com"}
	cache.EXPECT().Get(mock.Anything, int64(99)).Return(expected, nil)

	user, err := svc.GetUserByID(context.Background(), 99)
	require.NoError(t, err)
	require.Equal(t, expected.Name, user.Name)
}

func TestGetUserByID_CacheMiss_DBFallback(t *testing.T) {
	svc, repo, cache := newSvc(t)

	expected := &domain.User{ID: 5, Name: "DB User", Email: "db@example.com"}
	cache.EXPECT().Get(mock.Anything, int64(5)).Return(nil, errors.New("cache miss"))
	repo.EXPECT().GetByID(mock.Anything, int64(5)).Return(expected, nil)
	cache.EXPECT().Set(mock.Anything, expected).Return(nil)

	user, err := svc.GetUserByID(context.Background(), 5)
	require.NoError(t, err)
	require.Equal(t, "DB User", user.Name)
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc, repo, cache := newSvc(t)

	cache.EXPECT().Get(mock.Anything, int64(999)).Return(nil, errors.New("cache miss"))
	repo.EXPECT().GetByID(mock.Anything, int64(999)).Return(nil, errors.New("not found"))

	_, err := svc.GetUserByID(context.Background(), 999)
	require.Error(t, err)
}

func TestListUsers(t *testing.T) {
	svc, repo, _ := newSvc(t)

	filter := domain.ListFilter{Page: 1, Limit: 10}
	expected := &domain.UserList{
		Data:  []*domain.User{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}},
		Total: 2,
		Page:  1,
		Limit: 10,
	}
	repo.EXPECT().List(mock.Anything, filter).Return(expected, nil)

	result, err := svc.ListUsers(context.Background(), filter)
	require.NoError(t, err)
	require.Equal(t, int64(2), result.Total)
}

func TestUpdateUser_Success(t *testing.T) {
	svc, repo, cache := newSvc(t)

	newName := "New Name"
	input := domain.UpdateUserInput{Name: &newName}
	updated := &domain.User{ID: 1, Name: newName}

	repo.EXPECT().Update(mock.Anything, int64(1), input).Return(updated, nil)
	cache.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
	cache.EXPECT().Set(mock.Anything, updated).Return(nil)

	res, err := svc.UpdateUser(context.Background(), 1, input)
	require.NoError(t, err)
	require.Equal(t, newName, res.Name)
}

func TestDeleteUser_Success(t *testing.T) {
	svc, repo, cache := newSvc(t)

	repo.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
	cache.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)

	err := svc.DeleteUser(context.Background(), 1)
	require.NoError(t, err)
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc, repo, _ := newSvc(t)

	repo.EXPECT().Delete(mock.Anything, int64(999)).Return(errors.New("not found"))

	err := svc.DeleteUser(context.Background(), 999)
	require.Error(t, err)
}

// --- Email Service Tests ---

func TestCreateUser_WithEmailService(t *testing.T) {
	svc, repo, cache := newSvc(t)
	emailSvc := mocks.NewEmailService(t)
	svc.SetEmailService(emailSvc)

	input := domain.CreateUserInput{Name: "Email Test", Email: "test@example.com", Password: "pass"}
	created := &domain.User{ID: 1, Name: input.Name, Email: input.Email}

	repo.EXPECT().GetByEmail(mock.Anything, input.Email).Return(nil, errors.New("not found"))
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(created, nil)
	cache.EXPECT().Set(mock.Anything, created).Return(nil)
	emailSvc.EXPECT().SendWelcomeEmail(mock.Anything, created.Email, created.Name).Return(nil)

	_, err := svc.CreateUser(context.Background(), input)
	require.NoError(t, err)
}

// --- Repository Error Tests ---

func TestCreateUser_RepositoryError(t *testing.T) {
	svc, repo, _ := newSvc(t)

	repo.EXPECT().GetByEmail(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	_, err := svc.CreateUser(context.Background(), domain.CreateUserInput{Email: "test@test.com", Password: "password123"})
	require.Error(t, err)
}

// --- Cache Error Tests ---

func TestGetUserByID_CacheError_RecoverFromDB(t *testing.T) {
	svc, repo, cache := newSvc(t)

	expected := &domain.User{ID: 5, Name: "DB User"}
	cache.EXPECT().Get(mock.Anything, int64(5)).Return(nil, errors.New("cache fail"))
	repo.EXPECT().GetByID(mock.Anything, int64(5)).Return(expected, nil)
	cache.EXPECT().Set(mock.Anything, expected).Return(errors.New("cache set fail"))

	user, err := svc.GetUserByID(context.Background(), 5)
	require.NoError(t, err)
	require.Equal(t, "DB User", user.Name)
}

// --- Pagination Tests ---

func TestListUsers_DefaultPagination(t *testing.T) {
	svc, repo, _ := newSvc(t)

	repo.EXPECT().List(mock.Anything, mock.MatchedBy(func(f domain.ListFilter) bool {
		return f.Page == 1 && f.Limit == 10
	})).Return(&domain.UserList{}, nil)

	_, err := svc.ListUsers(context.Background(), domain.ListFilter{Page: 0, Limit: 0})
	require.NoError(t, err)
}
