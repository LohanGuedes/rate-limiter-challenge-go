package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestIsAllowed_Table(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		redisVal    string
		redisErr    error
		expectIncr  bool
		expectAllow bool
		expectErr   bool
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
			name:        "Key exists and at limit",
			redisVal:    "3",
			expectIncr:  false,
			expectAllow: false,
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
			limiter := New(client, 3, 60)
			key := "rate_limit:test_user"

			if tt.redisErr != nil {
				if tt.redisErr == redis.Nil {
					mock.ExpectGet(key).RedisNil()
				} else {
					mock.ExpectGet(key).SetErr(tt.redisErr)
				}
			} else {
				mock.ExpectGet(key).SetVal(tt.redisVal)
			}

			if tt.expectIncr {
				mock.ExpectTxPipeline()
				mock.ExpectIncr(key).SetVal(1)
				mock.ExpectExpireNX(key, 60*time.Second).SetVal(true)
				mock.ExpectTxPipelineExec()
			}

			allowed, err := limiter.IsAllowed(ctx, "test_user")

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if allowed != tt.expectAllow {
				t.Errorf("expected allowed = %v, got %v", tt.expectAllow, allowed)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet redis expectations: %v", err)
			}
		})
	}
}
