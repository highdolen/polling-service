package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, addr string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, closeErr
		}

		return nil, err
	}

	return client, nil
}
