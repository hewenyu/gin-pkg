package util

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient wraps Redis operations
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(host string, port int, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// BlacklistToken adds a token to the blacklist
func (r *RedisClient) BlacklistToken(tokenID string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	return r.client.Set(ctx, key, "1", expiration).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *RedisClient) IsTokenBlacklisted(tokenID string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("blacklist:token:%s", tokenID)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// StoreNonce stores a nonce with an expiration time
func (r *RedisClient) StoreNonce(nonce string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("nonce:%s", nonce)
	return r.client.Set(ctx, key, "1", expiration).Err()
}

// GetNonce checks if a nonce exists
func (r *RedisClient) GetNonce(nonce string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("nonce:%s", nonce)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// InvalidateNonce removes a nonce
func (r *RedisClient) InvalidateNonce(nonce string) error {
	ctx := context.Background()
	key := fmt.Sprintf("nonce:%s", nonce)
	return r.client.Del(ctx, key).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
