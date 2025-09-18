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
	limit      int
	windowSize int
	client     *redis.Client
}

// New creates a redis-based rate-limiter
func New(client *redis.Client, count, windowSize int) *RateLimiter {
	return &RateLimiter{count, windowSize, client}
}

func (rl *RateLimiter) IsAllowed(ctx context.Context, id string) (bool, error) {
	key := "rate_limit:" + id
	val, err := rl.client.Get(ctx, key).Result()
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

	if currentCount >= rl.limit {
		return false, nil
	}

	p := rl.client.TxPipeline()
	p.Incr(ctx, key)
	p.ExpireNX(ctx, key, time.Duration(rl.windowSize)*time.Second)
	_, err = p.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to atomic increment rate-limiter counter: %w", err)
	}

	return true, nil
}
