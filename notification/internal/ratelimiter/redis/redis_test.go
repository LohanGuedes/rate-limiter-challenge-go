package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
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
			limiter := New(client)

			id := "968af933-64e3-4890-bd3c-50158bdadf0c"
			key := model.NotificationTypeStatus.GenKey(id)
			redisKey := "rate_limit:" + key

			if tt.redisErr != nil {
				if errors.Is(tt.redisErr, redis.Nil) {
					mock.ExpectGet(redisKey).RedisNil()
				} else {
					mock.ExpectGet(redisKey).SetErr(tt.redisErr)
				}
			} else {
				mock.ExpectGet(redisKey).SetVal(tt.redisVal)
			}

			if tt.expectIncr {
				mock.ExpectTxPipeline()
				mock.ExpectIncr(redisKey).SetVal(1)
				mock.ExpectExpireNX(redisKey, 60*time.Second).SetVal(true)
				mock.ExpectTxPipelineExec()
			}

			allowed, err := limiter.IsAllowed(ctx, key, 3, 60)

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
