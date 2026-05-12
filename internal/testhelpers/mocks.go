package testhelpers

import (
	"context"
	"errors"
	"sync"
	"time"

	"go-usersvc-demo/internal/domain"

	"github.com/redis/go-redis/v9"
)

type MockUserRepository struct {
	mu         sync.RWMutex
	users      map[int64]*domain.User
	nextID     int64
	shouldFail bool
	failErr    error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	if m.shouldFail {
		return nil, m.failErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, existing := range m.users {
		if existing.Email == u.Email {
			return nil, errors.New("email already registered")
		}
	}

	u.ID = m.nextID
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	m.nextID++

	m.users[u.ID] = u
	return u, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.shouldFail {
		return nil, m.failErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.shouldFail {
		return nil, m.failErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) List(ctx context.Context, filter domain.ListFilter) (*domain.UserList, error) {
	if m.shouldFail {
		return nil, m.failErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}

	return &domain.UserList{
		Data:  users,
		Total: int64(len(users)),
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}

func (m *MockUserRepository) Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error) {
	if m.shouldFail {
		return nil, m.failErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

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
	if input.Password != nil {
		u.Password = *input.Password
	}
	u.UpdatedAt = time.Now()

	return u, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	if m.shouldFail {
		return m.failErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.users[id]; !ok {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) SetupError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = true
	m.failErr = err
}

func (m *MockUserRepository) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = false
	m.failErr = nil
}

func (m *MockUserRepository) GetAllUsers() []*domain.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []*domain.User
	for _, u := range m.users {
		users = append(users, u)
	}
	return users
}

type MockUserCache struct {
	mu    sync.RWMutex
	cache map[int64]*domain.User
}

func NewMockUserCache() *MockUserCache {
	return &MockUserCache{
		cache: make(map[int64]*domain.User),
	}
}

func (m *MockUserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, ok := m.cache[id]
	if !ok {
		return nil, redis.Nil
	}
	return u, nil
}

func (m *MockUserCache) Set(ctx context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cache[u.ID] = u
	return nil
}

func (m *MockUserCache) Delete(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, id)
	return nil
}
