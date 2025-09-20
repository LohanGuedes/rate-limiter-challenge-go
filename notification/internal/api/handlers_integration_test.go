package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/ratelimiter/redis"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/google/uuid"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedisContainer(t *testing.T) (*redisclient.Client, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	redisClient := redisclient.NewClient(&redisclient.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
		DB:   0,
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	cleanup := func() {
		redisClient.Close()
		redisContainer.Terminate(ctx)
	}

	return redisClient, cleanup
}

type realConfigProviderMock struct{}

func (r *realConfigProviderMock) GetConfig(nt model.NotificationType) (config.RLConfig, bool) {
	configs := map[model.NotificationType]config.RLConfig{
		model.NotificationTypeNews:      {Limit: 1, WindowSize: 86400}, // 1 per day
		model.NotificationTypeStatus:    {Limit: 2, WindowSize: 60},    // 2 per minute
		model.NotificationTypeMarketing: {Limit: 3, WindowSize: 3600},  // 3 per hour
	}
	cfg, ok := configs[nt]
	return cfg, ok
}

func TestIntegrationNewsNotificationRateLimit(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)
	configProvider := &realConfigProviderMock{}
	ctrl := notification.NewController(rateLimiter, configProvider)
	app := New(logger, redisClient, ctrl)

	userID := uuid.New()

	t.Run("first news message succeeds", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeNews,
			Message:          "This is the first news notification",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
	})

	t.Run("second news message fails", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeNews,
			Message:          "This is the second news notification",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusTooManyRequests, resp.Code, resp.Body.String())
		}

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		if response["message"] != "too many messages of that type sent" {
			t.Errorf("Expected rate limit message, got '%v'", response["message"])
		}
	})
}

func TestIntegrationStatusNotificationRateLimit(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)
	configProvider := &realConfigProviderMock{}
	ctrl := notification.NewController(rateLimiter, configProvider)
	app := New(logger, redisClient, ctrl)

	userID := uuid.New()

	for i := 1; i <= 2; i++ {
		t.Run(fmt.Sprintf("status message %d succeeds", i), func(t *testing.T) {
			notification := model.Notification{
				UserID:           userID,
				NotificationType: model.NotificationTypeStatus,
				Message:          fmt.Sprintf("Status update #%d", i),
			}

			resp := sendNotification(t, app, notification)
			if resp.Code != http.StatusCreated {
				t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
			}
		})
	}

	t.Run("third status message fails", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeStatus,
			Message:          "Status update #3 - should fail",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusTooManyRequests, resp.Code, resp.Body.String())
		}
	})
}

func TestIntegrationMarketingNotificationRateLimit(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)
	configProvider := &realConfigProviderMock{}
	ctrl := notification.NewController(rateLimiter, configProvider)
	app := New(logger, redisClient, ctrl)

	userID := uuid.New()

	for i := 1; i <= 3; i++ {
		t.Run(fmt.Sprintf("marketing message %d succeeds", i), func(t *testing.T) {
			notification := model.Notification{
				UserID:           userID,
				NotificationType: model.NotificationTypeMarketing,
				Message:          fmt.Sprintf("Marketing campaign #%d", i),
			}

			resp := sendNotification(t, app, notification)
			if resp.Code != http.StatusCreated {
				t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
			}
		})
	}

	t.Run("fourth marketing message fails", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeMarketing,
			Message:          "Marketing campaign #4 - should fail",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusTooManyRequests, resp.Code, resp.Body.String())
		}
	})
}

func TestIntegrationMultipleUsersIsolatedRateLimits(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)
	configProvider := &realConfigProviderMock{}
	ctrl := notification.NewController(rateLimiter, configProvider)
	app := New(logger, redisClient, ctrl)

	user1 := uuid.New()
	user2 := uuid.New()

	t.Run("user1 exhausts news limit", func(t *testing.T) {
		notification := model.Notification{
			UserID:           user1,
			NotificationType: model.NotificationTypeNews,
			Message:          "User 1 news",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}

		resp = sendNotification(t, app, notification)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, resp.Code)
		}
	})

	t.Run("user2 can still send", func(t *testing.T) {
		notification := model.Notification{
			UserID:           user2,
			NotificationType: model.NotificationTypeNews,
			Message:          "User 2 news",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
	})
}

func TestIntegrationMixedNotificationTypesIndependentLimits(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)
	configProvider := &realConfigProviderMock{}
	ctrl := notification.NewController(rateLimiter, configProvider)
	app := New(logger, redisClient, ctrl)

	userID := uuid.New()

	t.Run("exhaust news limit", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeNews,
			Message:          "News notification",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}
	})

	t.Run("status still works after news exhausted", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeStatus,
			Message:          "Status notification",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
	})

	t.Run("marketing still works after news exhausted", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeMarketing,
			Message:          "Marketing notification",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
	})
}

func sendNotification(t *testing.T, app *Application, notification model.Notification) *httptest.ResponseRecorder {
	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)
	return w
}

func TestIntegrationRateLimitWindowExpiration(t *testing.T) {
	redisClient, cleanup := setupRedisContainer(t)
	defer cleanup()

	logger := slog.Default()
	rateLimiter := redis.New(redisClient)

	// Override config provider
	shortWindowConfig := &testConfigProvider{
		configs: map[model.NotificationType]config.RLConfig{
			model.NotificationTypeStatus: {Limit: 1, WindowSize: 2}, // 1 per 2 seconds
		},
	}

	ctrl := notification.NewController(rateLimiter, shortWindowConfig)
	app := New(logger, redisClient, ctrl)

	userID := uuid.New()

	t.Run("first message succeeds", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeStatus,
			Message:          "First message",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}
	})

	t.Run("second message fails within window", func(t *testing.T) {
		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeStatus,
			Message:          "Second message",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, resp.Code)
		}
	})

	t.Run("message succeeds after window expires", func(t *testing.T) {
		time.Sleep(3 * time.Second) // Wait for 2-second or more window to expire

		notification := model.Notification{
			UserID:           userID,
			NotificationType: model.NotificationTypeStatus,
			Message:          "Third message after window expiry",
		}

		resp := sendNotification(t, app, notification)
		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d after window expiry, got %d. Body: %s",
				http.StatusCreated, resp.Code, resp.Body.String())
		}
	})
}

type testConfigProvider struct {
	configs map[model.NotificationType]config.RLConfig
}

func (t *testConfigProvider) GetConfig(nt model.NotificationType) (config.RLConfig, bool) {
	cfg, ok := t.configs[nt]
	return cfg, ok
}
