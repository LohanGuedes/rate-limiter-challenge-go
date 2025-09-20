package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/model"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/pkg/ratelimit"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type mockRateLimiter struct {
	isAllowedFunc func(ctx context.Context, key string, limit, windowSize int) (bool, error)
}

func (m *mockRateLimiter) IsAllowed(ctx context.Context, key string, limit, windowSize int) (bool, error) {
	if m.isAllowedFunc != nil {
		return m.isAllowedFunc(ctx, key, limit, windowSize)
	}
	return true, nil
}

type mockConfigProvider struct {
	configs map[model.NotificationType]config.RLConfig
}

func (m *mockConfigProvider) GetConfig(nt model.NotificationType) (config.RLConfig, bool) {
	cfg, ok := m.configs[nt]
	return cfg, ok
}

func newMockConfigProvider() *mockConfigProvider {
	return &mockConfigProvider{
		configs: map[model.NotificationType]config.RLConfig{
			model.NotificationTypeNews:      {Limit: 1, WindowSize: 86400},
			model.NotificationTypeStatus:    {Limit: 2, WindowSize: 60},
			model.NotificationTypeMarketing: {Limit: 3, WindowSize: 3600},
		},
	}
}

func TestHandleSendNotification_ValidRequest(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}

	mockRL := &mockRateLimiter{
		isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
			return true, nil // Allow request
		},
	}

	configProvider := newMockConfigProvider()
	ctrl := notification.NewController(mockRL, configProvider)
	app := New(logger, redisClient, ctrl)

	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: model.NotificationTypeNews,
		Message:          "This is a valid test message that is long enough",
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "Application/json" {
		t.Errorf("Expected Content-Type 'Application/json', got '%s'", contentType)
	}

	var response map[string]any
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Message Sent" {
		t.Errorf("Expected message 'Message Sent', got '%v'", response["message"])
	}
}

func TestHandleSendNotification_InvalidJSON(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}
	app := New(logger, redisClient, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/notify/send", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	// Should return 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleSendNotification_ValidationFailure(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}
	app := New(logger, redisClient, ctrl)

	// Create notification with validation errors
	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: "invalid-type", // Invalid type
		Message:          "short",        // Too short
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]any
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should have validation error fields
	if len(response) == 0 {
		t.Error("Expected validation errors in response")
	}
}

func TestHandleSendNotification_UnknownNotificationType(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}

	mockRL := &mockRateLimiter{}

	emptyConfigProvider := &mockConfigProvider{
		configs: map[model.NotificationType]config.RLConfig{},
	}

	ctrl := notification.NewController(mockRL, emptyConfigProvider)
	app := New(logger, redisClient, ctrl)

	// Create notification with unconfigured type
	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: model.NotificationTypeNews, // Not in empty config
		Message:          "This is a valid test message that is long enough",
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}

	var response map[string]any
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "this notification type handler was not found" {
		t.Errorf("Expected unknown type message, got '%v'", response["message"])
	}
}

func TestHandleSendNotification_RateLimiterError(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}

	mockRL := &mockRateLimiter{
		isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
			return false, errors.New("redis connection error")
		},
	}

	configProvider := newMockConfigProvider()
	ctrl := notification.NewController(mockRL, configProvider)
	app := New(logger, redisClient, ctrl)

	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: model.NotificationTypeNews,
		Message:          "This is a valid test message that is long enough",
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d got %d",
			http.StatusCreated, w.Code)
	}
}

func TestHandleSendNotification_AllNotificationTypes(t *testing.T) {
	testCases := []struct {
		name             string
		notificationType model.NotificationType
	}{
		{"news_notification", model.NotificationTypeNews},
		{"status_notification", model.NotificationTypeStatus},
		{"marketing_notification", model.NotificationTypeMarketing},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.Default()
			redisClient := &redis.Client{}

			mockRL := &mockRateLimiter{
				isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
					return true, nil
				},
			}

			configProvider := newMockConfigProvider()
			ctrl := notification.NewController(mockRL, configProvider)
			app := New(logger, redisClient, ctrl)

			validID := uuid.New()
			notification := model.Notification{
				UserID:           validID,
				NotificationType: tc.notificationType,
				Message:          "This is a valid test message that is long enough",
			}

			jsonPayload, err := json.Marshal(notification)
			if err != nil {
				t.Fatalf("Failed to marshal notification: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			app.handleSendNotification(w, req)

			if w.Code != http.StatusCreated {
				t.Errorf("Expected status code %d for %s, got %d. Body: %s",
					http.StatusCreated, tc.notificationType, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandleSendNotification_EmptyBody(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}
	app := New(logger, redisClient, ctrl)

	req := httptest.NewRequest(http.MethodPost, "/notify/send", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleSendNotification_MissingFields(t *testing.T) {
	testCases := []struct {
		name    string
		payload map[string]any
	}{
		{
			name: "missing_notification_type",
			payload: map[string]any{
				"userId":  uuid.New().String(),
				"message": "This is a valid test message that is long enough",
			},
		},
		{
			name: "missing_message",
			payload: map[string]any{
				"userId":           uuid.New().String(),
				"notificationType": "news-notification",
			},
		},
		{
			name: "invalid_user_id",
			payload: map[string]any{
				"userId":           "invalid-uuid",
				"notificationType": "news-notification",
				"message":          "This is a valid test message that is long enough",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.Default()
			redisClient := &redis.Client{}
			ctrl := &notification.Controller{}
			app := New(logger, redisClient, ctrl)

			jsonPayload, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			app.handleSendNotification(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status code %d for %s, got %d. Body: %s", http.StatusBadRequest, tc.name, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandleSendNotification_RateLimitExceededWithRetryAfter(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}

	retryAfterDuration := 45 * time.Second
	mockRL := &mockRateLimiter{
		isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
			return false, ratelimit.NewLimitExceededError(retryAfterDuration, "rate limit exceeded")
		},
	}

	configProvider := newMockConfigProvider()
	ctrl := notification.NewController(mockRL, configProvider)
	app := New(logger, redisClient, ctrl)

	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: model.NotificationTypeNews,
		Message:          "This is a valid test message that is long enough",
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.handleSendNotification(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status code %d, got %d. Body: %s", http.StatusTooManyRequests, w.Code, w.Body.String())
	}

	// Check Retry-After header
	retryAfter := w.Header().Get("Retry-After")
	expectedRetryAfter := strconv.Itoa(int(retryAfterDuration.Seconds()))
	if retryAfter != expectedRetryAfter {
		t.Errorf("Expected Retry-After header '%s', got '%s'", expectedRetryAfter, retryAfter)
	}

	var response map[string]any
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "too many messages of that type sent" {
		t.Errorf("Expected rate limit message, got '%v'", response["message"])
	}
}

func TestHandleSendNotification_RateLimitVariousRetryAfterValues(t *testing.T) {
	testCases := []struct {
		name           string
		retryAfter     time.Duration
		expectedHeader string
	}{
		{"30_seconds", 30 * time.Second, "30"},
		{"1_minute", 60 * time.Second, "60"},
		{"5_minutes", 5 * time.Minute, "300"},
		{"1_hour", time.Hour, "3600"},
		{"fractional_seconds", 45500 * time.Millisecond, "46"}, // Rounds up
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.Default()
			redisClient := &redis.Client{}

			mockRL := &mockRateLimiter{
				isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
					return false, ratelimit.NewLimitExceededError(tc.retryAfter, "rate limit exceeded")
				},
			}

			configProvider := newMockConfigProvider()
			ctrl := notification.NewController(mockRL, configProvider)
			app := New(logger, redisClient, ctrl)

			validID := uuid.New()
			notification := model.Notification{
				UserID:           validID,
				NotificationType: model.NotificationTypeStatus,
				Message:          "Test retry after values",
			}

			jsonPayload, err := json.Marshal(notification)
			if err != nil {
				t.Fatalf("Failed to marshal notification: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			app.handleSendNotification(w, req)

			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, w.Code)
			}

			retryAfter := w.Header().Get("Retry-After")
			if retryAfter != tc.expectedHeader {
				t.Errorf("Expected Retry-After header '%s', got '%s'", tc.expectedHeader, retryAfter)
			}
		})
	}
}

func BenchmarkHandleSendNotification(b *testing.B) {
	logger := slog.Default()
	redisClient := &redis.Client{}

	mockRL := &mockRateLimiter{
		isAllowedFunc: func(ctx context.Context, key string, limit, windowSize int) (bool, error) {
			return true, nil
		},
	}

	configProvider := newMockConfigProvider()
	ctrl := notification.NewController(mockRL, configProvider)
	app := New(logger, redisClient, ctrl)

	validID := uuid.New()
	notification := model.Notification{
		UserID:           validID,
		NotificationType: model.NotificationTypeNews,
		Message:          "This is a valid test message that is long enough for benchmarking",
	}

	jsonPayload, err := json.Marshal(notification)
	if err != nil {
		b.Fatalf("Failed to marshal notification: %v", err)
	}

	for b.Loop() {
		req := httptest.NewRequest(http.MethodPost, "/notify/send", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		app.handleSendNotification(w, req)
	}
}
