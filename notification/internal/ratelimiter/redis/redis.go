package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter defines a redis-based rate-limiter
type RateLimiter struct {
	client *redis.Client
}

// New creates a redis-based rate-limiter
func New(client *redis.Client) *RateLimiter {
	return &RateLimiter{client}
}

func (rl *RateLimiter) IsAllowed(ctx context.Context, key string, limit, windowSize int) (bool, error) {
	rediskey := "rate_limit:" + key
	val, err := rl.client.Get(ctx, rediskey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	currentCount := 0
	if err == nil {
		currentCount, err = strconv.Atoi(val)
		if err != nil {
			return false, fmt.Errorf("failed to convert value to int: %w", err)
		}
	}

	if currentCount >= limit {
		return false, nil
	}

	p := rl.client.TxPipeline()
	p.Incr(ctx, rediskey)
	p.ExpireNX(ctx, rediskey, time.Duration(windowSize)*time.Second)
	_, err = p.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to atomic increment rate-limiter counter: %w", err)
	}

	return true, nil
}
