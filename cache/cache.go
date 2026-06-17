package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/laura838/shld-i-buy-this/models"
	"github.com/redis/go-redis/v9"
)

func NewClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func GetCachedDecision(ctx context.Context, client *redis.Client, key string) (*models.Decision, error) {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var d models.Decision
	if err := json.Unmarshal([]byte(val), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func SetCachedDecision(ctx context.Context, client *redis.Client, key string, decision *models.Decision, ttl time.Duration) error {
	b, err := json.Marshal(decision)
	if err != nil {
		return err
	}
	return client.Set(ctx, key, b, ttl).Err()
}

// SetCooldown sets a cooldown lock using SETNX, storing the verdict. Only sets if key does not exist.
func SetCooldown(ctx context.Context, client *redis.Client, key string, verdict string, ttl time.Duration) error {
	return client.SetNX(ctx, key, verdict, ttl).Err()
}

// CheckCooldown returns the stored verdict if the product is in cooldown, or ("", nil) if not.
func CheckCooldown(ctx context.Context, client *redis.Client, key string) (string, error) {
	val, err := client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// CheckRateLimit increments the request counter and returns true if under limit.
func CheckRateLimit(ctx context.Context, client *redis.Client, key string, limit int, window time.Duration) (bool, error) {
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		client.Expire(ctx, key, window)
	}
	return count <= int64(limit), nil
}
