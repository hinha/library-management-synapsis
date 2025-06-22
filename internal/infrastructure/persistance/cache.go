package persistance

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hinha/library-management-synapsis/cmd/config"
	"strconv"
)

func NewRedisConnection(cfg config.ServiceConfig) (*redis.Client, error) {
	dbToken, err := strconv.Atoi(cfg.CacheDbToken)
	if err != nil {
		return nil, fmt.Errorf("invalid CacheDbToken: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.CacheHost + ":" + cfg.CachePort,
		Password: cfg.CachePassword,
		DB:       dbToken,
	})

	// Test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
