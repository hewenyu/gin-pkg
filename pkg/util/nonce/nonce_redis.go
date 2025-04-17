package nonce

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type NonceRedisService struct {
	client   *redis.Client
	redisKey string
}

func NewNonceRedisService(redisClient *redis.Client) *NonceRedisService {
	return &NonceRedisService{
		client:   redisClient,
		redisKey: "nonce",
	}
}

// StoreNonce stores a nonce with an expiration time
func (r *NonceRedisService) StoreNonce(nonce string, expiration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", r.redisKey, nonce)
	return r.client.Set(ctx, key, "1", expiration).Err()
}

// GetNonce checks if a nonce exists and invalidates it if found
func (r *NonceRedisService) GetNonce(nonce string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:%s", r.redisKey, nonce)

	// Use a transaction to check and delete in one atomic operation
	txf := func(tx *redis.Tx) error {
		// Check if the key exists
		exists, err := tx.Exists(ctx, key).Result()
		if err != nil {
			return err
		}

		// If the key exists, delete it to ensure one-time use
		if exists > 0 {
			_, err = tx.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}

		return nil
	}

	// Execute the transaction
	err := r.client.Watch(ctx, txf, key)
	if err != nil {
		return false, fmt.Errorf("transaction failed: %w", err)
	}

	return true, nil
}

// Close closes the Redis connection
func (r *NonceRedisService) Close() error {
	return r.client.Close()
}
