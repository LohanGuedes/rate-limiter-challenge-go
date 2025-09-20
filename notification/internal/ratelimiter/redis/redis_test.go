package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/ratelimit"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestIsAllowed_Table(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		redisVal        string
		redisErr        error
		ttlDuration     time.Duration
		ttlErr          error
		expectIncr      bool
		expectAllow     bool
		expectErr       bool
		expectRateLimit bool
	}{
		{
			name:        "First request - key does not exist",
			redisErr:    redis.Nil,
			expectIncr:  true,
			expectAllow: true,
		},
		{
			name:        "Key exists and below limit",
			redisVal:    "2",
			expectIncr:  true,
			expectAllow: true,
		},
		{
			name:            "Key exists and at limit - with TTL",
			redisVal:        "3",
			ttlDuration:     45 * time.Second,
			expectIncr:      false,
			expectAllow:     false,
			expectRateLimit: true,
		},
		{
			name:            "Key exists and at limit - TTL error fallback",
			redisVal:        "3",
			ttlErr:          errors.New("ttl failed"),
			expectIncr:      false,
			expectAllow:     false,
			expectRateLimit: true,
		},
		{
			name:        "Invalid redis value",
			redisVal:    "not-a-number",
			expectErr:   true,
			expectAllow: false,
		},
		{
			name:        "Redis GET returns unexpected error",
			redisErr:    errors.New("connection dropped"),
			expectErr:   true,
			expectAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			limiter := New(client)

			id := "968af933-64e3-4890-bd3c-50158bdadf0c"
			key := model.NotificationTypeStatus.GenKey(id)
			redisKey := "rate_limit:" + key
			windowSize := 60

			if tt.redisErr != nil {
				if errors.Is(tt.redisErr, redis.Nil) {
					mock.ExpectGet(redisKey).RedisNil()
				} else {
					mock.ExpectGet(redisKey).SetErr(tt.redisErr)
				}
			} else {
				mock.ExpectGet(redisKey).SetVal(tt.redisVal)
			}

			// If rate limited, expect TTL call
			if tt.expectRateLimit {
				if tt.ttlErr != nil {
					mock.ExpectTTL(redisKey).SetErr(tt.ttlErr)
				} else {
					mock.ExpectTTL(redisKey).SetVal(tt.ttlDuration)
				}
			}

			if tt.expectIncr {
				mock.ExpectTxPipeline()
				mock.ExpectIncr(redisKey).SetVal(1)
				mock.ExpectExpireNX(redisKey, time.Duration(windowSize)*time.Second).SetVal(true)
				mock.ExpectTxPipelineExec()
			}

			allowed, err := limiter.IsAllowed(ctx, key, 3, windowSize)

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got none")
			}
			if !tt.expectErr && err != nil && !tt.expectRateLimit {
				t.Errorf("unexpected error: %v", err)
			}
			if allowed != tt.expectAllow {
				t.Errorf("expected allowed = %v, got %v", tt.expectAllow, allowed)
			}

			// Check rate limit error specifics
			if tt.expectRateLimit {
				var rateLimitErr *ratelimit.LimitExceededError
				if !errors.As(err, &rateLimitErr) {
					t.Errorf("expected LimitExceededError, got %T: %v", err, err)
				} else {
					expectedRetryAfter := tt.ttlDuration
					if tt.ttlErr != nil {
						expectedRetryAfter = time.Duration(windowSize) * time.Second
					}

					if rateLimitErr.RetryAfter != expectedRetryAfter {
						t.Errorf("expected RetryAfter %v, got %v", expectedRetryAfter, rateLimitErr.RetryAfter)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet redis expectations: %v", err)
			}
		})
	}
}

func TestIsAllowed_RetryAfterCalculation(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	limiter := New(client)

	id := "test-user"
	key := model.NotificationTypeNews.GenKey(id)
	redisKey := "rate_limit:" + key

	// Mock rate limit exceeded scenario
	mock.ExpectGet(redisKey).SetVal("5") // Above limit of 3
	mock.ExpectTTL(redisKey).SetVal(30 * time.Second)

	allowed, err := limiter.IsAllowed(ctx, key, 3, 60)

	if allowed {
		t.Error("expected request to be denied")
	}

	var rateLimitErr *ratelimit.LimitExceededError
	if !errors.As(err, &rateLimitErr) {
		t.Fatalf("expected LimitExceededError, got %T: %v", err, err)
	}

	if rateLimitErr.RetryAfter != 30*time.Second {
		t.Errorf("expected RetryAfter 30s, got %v", rateLimitErr.RetryAfter)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet redis expectations: %v", err)
	}
}
