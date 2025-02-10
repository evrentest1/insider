package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const timeout = 2 * time.Second

type Cache struct {
	Client *redis.Client
}

func New(ctx context.Context, address string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: address,
	})

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Cache{Client: client}, nil
}

func (c *Cache) IsOK(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := c.Client.Ping(ctx).Result()
	return err == nil
}
