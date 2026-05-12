package service_test

import (
	"context"
	"errors"
	"fmt"
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

// --- Mock Email Service ---

type mockEmailService struct {
	sentEmails map[string]int // email -> count
	shouldFail bool
}

func newMockEmailService() *mockEmailService {
	return &mockEmailService{sentEmails: make(map[string]int)}
}

func (m *mockEmailService) SendWelcomeEmail(ctx context.Context, email string, name string) error {
	if m.shouldFail {
		return errors.New("email service failed")
	}
	m.sentEmails[email]++
	return nil
}

// --- Integration Tests ---

func TestCreateUser_WithEmailService(t *testing.T) {
	svc, _, _ := newSvc()
	emailSvc := newMockEmailService()
	svc.SetEmailService(emailSvc)

	user, err := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Email Test",
		Email:    "emailtest@example.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
}

func TestCreateUser_WithoutEmailService(t *testing.T) {
	svc, _, _ := newSvc()

	user, err := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "No Email Service Test",
		Email:    "noemailsvc@example.com",
		Password: "pass123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID == 0 {
		t.Error("expected non-zero user ID")
	}
	if user.Name != "No Email Service Test" {
		t.Errorf("expected name 'No Email Service Test', got '%s'", user.Name)
	}
}

func TestUpdateUser_CacheInvalidation(t *testing.T) {
	svc, _, cache := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Cache Test",
		Email:    "cachetest@example.com",
		Password: "pass123",
	})

	cached, err := cache.Get(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("expected cached user, got error: %v", err)
	}
	if cached.Name != "Cache Test" {
		t.Errorf("expected cached name 'Cache Test', got '%s'", cached.Name)
	}

	newName := "Updated Name"
	updated, _ := svc.UpdateUser(context.Background(), user.ID, domain.UpdateUserInput{Name: &newName})

	cached, _ = cache.Get(context.Background(), user.ID)
	if cached.Name != "Updated Name" {
		t.Errorf("expected cached name 'Updated Name', got '%s'", cached.Name)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("expected updated name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestDeleteUser_CacheEviction(t *testing.T) {
	svc, _, cache := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Delete Cache Test",
		Email:    "delcachetest@example.com",
		Password: "pass123",
	})

	_, err := cache.Get(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("user should be in cache, got error: %v", err)
	}
	svc.DeleteUser(context.Background(), user.ID)

	_, err = cache.Get(context.Background(), user.ID)
	if err == nil {
		t.Fatal("expected user to be evicted from cache")
	}
}

func TestGetByEmail_Integration(t *testing.T) {
	svc, repo, _ := newSvc()

	input := domain.CreateUserInput{
		Name:     "Email Lookup",
		Email:    "emaillookup@example.com",
		Password: "pass123",
	}
	svc.CreateUser(context.Background(), input)

	user, err := repo.GetByEmail(context.Background(), "emaillookup@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Email != "emaillookup@example.com" {
		t.Errorf("expected email 'emaillookup@example.com', got '%s'", user.Email)
	}
}

// --- Mock Repository Error Tests ---

type errorMockRepo struct{}

func (m *errorMockRepo) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	return nil, errors.New("database error: create failed")
}

func (m *errorMockRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return nil, errors.New("database error: read failed")
}

func (m *errorMockRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, errors.New("database error: query failed")
}

func (m *errorMockRepo) List(ctx context.Context, filter domain.ListFilter) (*domain.UserList, error) {
	return nil, errors.New("database error: list failed")
}

func (m *errorMockRepo) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	return nil, errors.New("database error: update failed")
}

func (m *errorMockRepo) Delete(ctx context.Context, id int64) error {
	return errors.New("database error: delete failed")
}

func TestCreateUser_RepositoryError(t *testing.T) {
	svc := service.NewUserService(&errorMockRepo{}, newMockCache())

	_, err := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Error Test",
		Email:    "errortest@example.com",
		Password: "pass123",
	})
	if err == nil {
		t.Fatal("expected error from failing repository")
	}
}

func TestUpdateUser_RepositoryError(t *testing.T) {
	svc := service.NewUserService(&errorMockRepo{}, newMockCache())

	newName := "Updated"
	_, err := svc.UpdateUser(context.Background(), 1, domain.UpdateUserInput{Name: &newName})
	if err == nil {
		t.Fatal("expected error from failing repository")
	}
}

func TestDeleteUser_RepositoryError(t *testing.T) {
	svc := service.NewUserService(&errorMockRepo{}, newMockCache())

	err := svc.DeleteUser(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from failing repository")
	}
}

func TestGetUserByID_RepositoryError(t *testing.T) {
	svc := service.NewUserService(&errorMockRepo{}, newMockCache())

	_, err := svc.GetUserByID(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from failing repository")
	}
}

// --- Mock Cache Error Tests ---

type errorMockCache struct{}

func (c *errorMockCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	return nil, errors.New("cache error: connection failed")
}

func (c *errorMockCache) Set(ctx context.Context, u *domain.User) error {
	return errors.New("cache error: write failed")
}

func (c *errorMockCache) Delete(ctx context.Context, id int64) error {
	return errors.New("cache error: delete failed")
}

func TestGetUserByID_CacheError_RecoverFromDB(t *testing.T) {
	repo := newMockRepo()
	repo.users[5] = &domain.User{
		ID:        5,
		Name:      "DB User",
		Email:     "db@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	svc := service.NewUserService(repo, &errorMockCache{})

	user, err := svc.GetUserByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Name != "DB User" {
		t.Errorf("expected 'DB User', got '%s'", user.Name)
	}
}

// --- Pagination Tests ---

func TestListUsers_Pagination(t *testing.T) {
	svc, _, _ := newSvc()

	for i := 1; i <= 5; i++ {
		svc.CreateUser(context.Background(), domain.CreateUserInput{
			Name:     fmt.Sprintf("User%d", i),
			Email:    fmt.Sprintf("user%d@test.com", i),
			Password: "pass123",
		})
	}

	result1, err := svc.ListUsers(context.Background(), domain.ListFilter{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result1.Total != 5 {
		t.Errorf("expected total 5, got %d", result1.Total)
	}
}

func TestListUsers_DefaultPagination(t *testing.T) {
	svc, _, _ := newSvc()

	svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Test",
		Email:    "test@test.com",
		Password: "pass123",
	})

	result, err := svc.ListUsers(context.Background(), domain.ListFilter{Page: 0, Limit: 0})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
	if result.Limit != 10 {
		t.Errorf("expected limit 10, got %d", result.Limit)
	}
}

// --- Password Hashing Tests ---

func TestCreateUser_PasswordHashing(t *testing.T) {
	svc, repo, _ := newSvc()

	plainPassword := "mySecurePassword123"
	svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Hash Test",
		Email:    "hashtest@example.com",
		Password: plainPassword,
	})

	user, _ := repo.GetByEmail(context.Background(), "hashtest@example.com")
	if user.Password == plainPassword {
		t.Fatal("expected password to be hashed, but it matches plaintext")
	}
	if len(user.Password) == 0 {
		t.Fatal("expected hashed password to be non-empty")
	}
}

func TestUpdateUser_PasswordHashing(t *testing.T) {
	svc, repo, _ := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Pass Update",
		Email:    "passupd@example.com",
		Password: "oldpass123",
	})

	newPass := "newpass456"
	svc.UpdateUser(context.Background(), user.ID, domain.UpdateUserInput{Password: &newPass})

	updated, _ := repo.GetByID(context.Background(), user.ID)
	if updated.Password == newPass {
		t.Fatal("expected password to be hashed after update")
	}
}

// --- Complex Integration Test ---

func TestCompleteUserJourney(t *testing.T) {
	svc, repo, cache := newSvc()

	createInput := domain.CreateUserInput{
		Name:     "Journey User",
		Email:    "journey@example.com",
		Password: "securepass123",
	}
	user, err := svc.CreateUser(context.Background(), createInput)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	cached, err := cache.Get(context.Background(), user.ID)
	if err != nil || cached == nil {
		t.Fatal("expected user in cache after create")
	}
	retrieved, err := svc.GetUserByID(context.Background(), user.ID)
	if err != nil || retrieved.ID != user.ID {
		t.Fatalf("getByID failed: %v", err)
	}

	newName := "Updated Journey User"
	updated, err := svc.UpdateUser(context.Background(), user.ID, domain.UpdateUserInput{Name: &newName})
	if err != nil || updated.Name != newName {
		t.Fatalf("update failed: %v", err)
	}

	cachedAfterUpdate, _ := cache.Get(context.Background(), user.ID)
	if cachedAfterUpdate.Name != newName {
		t.Error("cache not updated after user update")
	}
	err = svc.DeleteUser(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, cacheErr := cache.Get(context.Background(), user.ID)
	if cacheErr == nil {
		t.Error("expected user to be evicted from cache")
	}

	_, repoErr := repo.GetByID(context.Background(), user.ID)
	if repoErr == nil {
		t.Error("expected user to be deleted from repository")
	}

}

// --- Concurrent Access Test ---

func TestConcurrentUserCreation(t *testing.T) {
	svc, _, _ := newSvc()
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_, err := svc.CreateUser(context.Background(), domain.CreateUserInput{
				Name:     fmt.Sprintf("Concurrent User %d", id),
				Email:    fmt.Sprintf("concurrent%d@example.com", id),
				Password: "pass123",
			})
			if err != nil {
				t.Errorf("concurrent create failed: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestConcurrentUserAccess(t *testing.T) {
	svc, _, _ := newSvc()

	user, _ := svc.CreateUser(context.Background(), domain.CreateUserInput{
		Name:     "Concurrent Access",
		Email:    "concurrent@example.com",
		Password: "pass123",
	})

	numGoroutines := 20
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := svc.GetUserByID(context.Background(), user.ID)
			if err != nil {
				t.Errorf("concurrent get failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
