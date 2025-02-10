package messagecache

import (
	"github.com/redis/go-redis/v9"
)

type Store struct {
	redis *redis.Client
}

func NewStore(r *redis.Client) *Store {
	return &Store{
		redis: r,
	}
}
