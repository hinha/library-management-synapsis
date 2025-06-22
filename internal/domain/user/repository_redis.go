package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/hinha/library-management-synapsis/cmd/config"
	"github.com/hinha/library-management-synapsis/internal/domain"
	"time"
)

//go:generate mockery --name=ICacheRepository --output=mocks --outpkg=mocks
type ICacheRepository interface {
	Ping(ctx context.Context) error
	SaveUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, id uint) (user *domain.User, err error)
}

// RedisClientInterface defines the Redis client methods used by CacheRepository
type RedisClientInterface interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type CacheRepository struct {
	client RedisClientInterface
}

func NewCacheRepository(client *redis.Client) ICacheRepository {
	return &CacheRepository{
		client: client,
	}
}

func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *CacheRepository) SaveUser(ctx context.Context, user *domain.User) error {
	data, _ := json.Marshal(user)
	return r.client.Set(ctx, config.RedisKeyUserPrefix+fmt.Sprintf("%d", user.ID), data, config.JwtTokenExpiration).Err()
}

func (r *CacheRepository) GetUser(ctx context.Context, id uint) (user *domain.User, err error) {
	data, err := r.client.Get(ctx, config.RedisKeyUserPrefix+fmt.Sprintf("%d", id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrUserNotFound
		}
		return nil, err // Other error
	}

	_ = json.Unmarshal(data, &user)
	return user, nil
}
