package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-usersvc-demo/internal/domain"
	"go-usersvc-demo/internal/service"
)

// --- Mock Repository ---

type mockRepo struct {
	users  map[int64]*domain.User
	nextID int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[int64]*domain.User), nextID: 1}
}

func (m *mockRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	for _, existing := range m.users {
		if existing.Email == u.Email {
			return nil, errors.New("email already exists")
		}
	}
	u.ID = m.nextID
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	m.nextID++
	m.users[u.ID] = u
	return u, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockRepo) List(ctx context.Context, filter domain.ListFilter) (*domain.UserList, error) {
	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return &domain.UserList{Data: users, Total: int64(len(users)), Page: filter.Page, Limit: filter.Limit}, nil
}

func (m *mockRepo) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	if input.Name != nil {
		u.Name = *input.Name
	}
	if input.Email != nil {
		u.Email = *input.Email
	}
	u.UpdatedAt = time.Now()
	return u, nil
}

func (m *mockRepo) Delete(ctx context.Context, id int64) error {
	if _, ok := m.users[id]; !ok {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

// --- Mock Cache ---

type mockCache struct {
	data map[int64]*domain.User
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[int64]*domain.User)}
}

func (c *mockCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	if u, ok := c.data[id]; ok {
		return u, nil
	}
	return nil, errors.New("cache miss")
}

func (c *mockCache) Set(ctx context.Context, u *domain.User) error {
	c.data[u.ID] = u
	return nil
}

func (c *mockCache) Delete(ctx context.Context, id int64) error {
	delete(c.data, id)
	return nil
}

// --- Tests ---

func newSvc() (*service.UserService, *mockRepo, *mockCache) {
	repo := newMockRepo()
	cache := newMockCache()
	svc := service.NewUserService(repo, cache)
	return svc, repo, cache
}

func TestCreateUser_Success(t *testing.T) {
	svc, _, _ := newSvc()

	user, err := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name: "John Doe", Email: "john@example.com", Password: "pass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if user.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got '%s'", user.Name)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	svc, _, _ := newSvc()

	input := domain.CreateUserInput{Name: "John", Email: "dup@example.com", Password: "pass123"}
	if _, err := svc.CreateUser(context.Background(), input); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := svc.CreateUser(context.Background(), input)
	if err == nil {
		t.Fatal("expected duplicate email error, got nil")
	}
}

func TestGetUserByID_CacheHit(t *testing.T) {
	svc, _, cache := newSvc()

	// Manually seed cache
	expected := &domain.User{ID: 99, Name: "Cached", Email: "cache@example.com"}
	_ = cache.Set(context.Background(), expected)

	user, err := svc.GetUserByID(context.Background(), 99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != "Cached" {
		t.Errorf("expected cached user, got %+v", user)
	}
}

func TestGetUserByID_CacheMiss_DBFallback(t *testing.T) {
	svc, repo, _ := newSvc()

	// Seed the repo directly
	repo.users[5] = &domain.User{ID: 5, Name: "DB User", Email: "db@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	user, err := svc.GetUserByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != "DB User" {
		t.Errorf("expected 'DB User', got '%s'", user.Name)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	svc, _, _ := newSvc()

	_, err := svc.GetUserByID(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestListUsers(t *testing.T) {
	svc, _, _ := newSvc()

	_, _ = svc.CreateUser(context.Background(), domain.CreateUserInput{Name: "A", Email: "a@test.com", Password: "pass123"})
	_, _ = svc.CreateUser(context.Background(), domain.CreateUserInput{Name: "B", Email: "b@test.com", Password: "pass123"})

	result, err := svc.ListUsers(context.Background(), domain.ListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 2 {
		t.Errorf("expected 2 users, got %d", result.Total)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	svc, _, _ := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{Name: "Old Name", Email: "upd@test.com", Password: "pass123"})

	newName := "New Name"
	updated, err := svc.UpdateUser(context.Background(), user.ID, domain.UpdateUserInput{Name: &newName})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected 'New Name', got '%s'", updated.Name)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	svc, _, _ := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{Name: "Del", Email: "del@test.com", Password: "pass123"})

	if err := svc.DeleteUser(context.Background(), user.ID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err := svc.GetUserByID(context.Background(), user.ID)
	if err == nil {
		t.Fatal("expected error after delete, user still exists")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	svc, _, _ := newSvc()

	err := svc.DeleteUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
