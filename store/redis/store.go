package redis

import (
	"bl-shifts/store"
	"context"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
)

type redisStore struct {
	rdb *redis.Client
}

func NewStore(address string) store.Store {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: os.Getenv("REDIS_PASSWORD"), // use the password from environment
		DB:       0,                            // default DB
	})
	return &redisStore{
		rdb: rdb,
	}
}

func (s *redisStore) FilterAndSaveCodes(ctx context.Context, codes []string) ([]string, error) {
	codesToSend := []string{}
	for _, code := range codes {
		exists, err := s.rdb.SIsMember(ctx, "shift_codes", code).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to check Redis for code %s: %w", code, err)
		}
		if !exists {
			codesToSend = append(codesToSend, code)
			s.rdb.SAdd(ctx, "shift_codes", code)
		}
	}

	return codesToSend, nil
}
