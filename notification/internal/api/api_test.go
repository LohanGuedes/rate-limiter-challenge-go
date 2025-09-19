package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	"github.com/redis/go-redis/v9"
)

func TestNew(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}

	app := New(logger, redisClient, ctrl)

	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}

	if app.Logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if app.RedisClient != redisClient {
		t.Error("Expected redis client to be set correctly")
	}

	if app.ctrl != ctrl {
		t.Error("Expected controller to be set correctly")
	}

	if app.Router == nil {
		t.Error("Expected router to be initialized")
	}
}

func TestBindRoutes(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}

	app := New(logger, redisClient, ctrl)
	handler := app.bindRoutes()

	if handler == nil {
		t.Fatal("Expected handler to be returned, got nil")
	}

	if handler != app.Router {
		t.Error("Expected handler to be the same as app.Router")
	}
}

func TestRouting(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}

	app := New(logger, redisClient, ctrl)
	server := httptest.NewServer(app.bindRoutes())
	defer server.Close()

	resp, err := http.Post(server.URL+"/notify/send", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Error("Expected route to exist, got 404")
	}
}

func TestHTTPMethods(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}

	app := New(logger, redisClient, ctrl)
	server := httptest.NewServer(app.bindRoutes())
	defer server.Close()

	// Test different HTTP methods
	methods := []struct {
		method         string
		expectedStatus int
	}{
		{"GET", http.StatusMethodNotAllowed},
		{"PUT", http.StatusMethodNotAllowed},
		{"DELETE", http.StatusMethodNotAllowed},
		{"PATCH", http.StatusMethodNotAllowed},
		{"POST", http.StatusBadRequest}, // POST is allowed but returns 400 for empty body
	}

	for _, m := range methods {
		req, err := http.NewRequest(m.method, server.URL+"/notify/send", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != m.expectedStatus {
			t.Errorf("Expected status %d for %s method, got %d", m.expectedStatus, m.method, resp.StatusCode)
		}
	}
}

func TestNonExistentRoute(t *testing.T) {
	logger := slog.Default()
	redisClient := &redis.Client{}
	ctrl := &notification.Controller{}

	app := New(logger, redisClient, ctrl)
	server := httptest.NewServer(app.bindRoutes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return 404 for non-existent routes
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent route, got %d", resp.StatusCode)
	}
}
