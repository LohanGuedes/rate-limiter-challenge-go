package api

import (
	"log/slog"
	"net/http"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
)

// Application defines the
type Application struct {
	Logger      *slog.Logger
	Router      *chi.Mux
	RedisClient *redis.Client
	ctrl        *notification.Controller
}

// New creates a HTTP Application for notification service
func New(logger *slog.Logger, redisClient *redis.Client, ctrl *notification.Controller) *Application {
	return &Application{
		Logger:      logger,
		Router:      chi.NewMux(),
		RedisClient: redisClient,
		ctrl:        ctrl,
	}
}

func (api *Application) bindRoutes() http.Handler {
	api.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	// Considering that this API can only be called within the the same
	// Network, therefore we are *NOT dealing with authentication on this api.
	// focusing on the rate-limiting when messaging.
	api.Router.Get("/healthcheck", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Healthy")) }))

	api.Router.Route("/notify", func(r chi.Router) {
		r.Post("/send", http.HandlerFunc(api.handleSendNotification))
	})

	return api.Router
}

// Start starts the Application on port 8080 and returns an error, if occurs
func (api *Application) Start() error {
	if err := http.ListenAndServe("localhost:8080", api.bindRoutes()); err != nil {
		return err
	}
	return nil
}
