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

type UserCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewUserCache(client *redis.Client) *UserCache {
	return &UserCache{client: client, ttl: defaultTTL}
}

func userKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

func (c *UserCache) Get(ctx context.Context, id int64) (*domain.User, error) {
	data, err := c.client.Get(ctx, userKey(id)).Bytes()
	if err != nil {
		return nil, err
	}

	user := &domain.User{}
	if err := json.Unmarshal(data, user); err != nil {
		return nil, fmt.Errorf("redis: unmarshal user: %w", err)
	}

	return user, nil
}

func (c *UserCache) Set(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("redis: marshal user: %w", err)
	}

	return c.client.Set(ctx, userKey(user.ID), data, c.ttl).Err()
}

func (c *UserCache) Delete(ctx context.Context, id int64) error {
	return c.client.Del(ctx, userKey(id)).Err()
}
