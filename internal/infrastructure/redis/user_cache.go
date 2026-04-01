package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-usersvc-demo/internal/domain"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 5 * time.Minute

// UserCache implements domain.UserCache using Redis.
type UserCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewUserCache creates a new UserCache.
func NewUserCache(client *redis.Client) *UserCache {
	return &UserCache{client: client, ttl: defaultTTL}
}

func userKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

// Get retrieves a user from cache by ID.
func (c *UserCache) Get(id int64) (*domain.User, error) {
	data, err := c.client.Get(context.Background(), userKey(id)).Bytes()
	if err != nil {
		return nil, err // redis.Nil if not found
	}

	user := &domain.User{}
	if err := json.Unmarshal(data, user); err != nil {
		return nil, fmt.Errorf("redis: unmarshal user: %w", err)
	}

	return user, nil
}

// Set stores a user in cache.
func (c *UserCache) Set(user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("redis: marshal user: %w", err)
	}

	return c.client.Set(context.Background(), userKey(user.ID), data, c.ttl).Err()
}

// Delete removes a user from cache.
func (c *UserCache) Delete(id int64) error {
	return c.client.Del(context.Background(), userKey(id)).Err()
}
