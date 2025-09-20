package main

import (
	"log/slog"
	"os"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/api"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	rlredis "github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/ratelimiter/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	defer client.Close()

	rateLimiter := rlredis.New(client)

	configs, err := config.LoadFromEmbedded()
	if err != nil {
		panic(err)
	}
	cfgProvider := config.NewRLConfigProvider(configs)
	ctrl := notification.NewController(rateLimiter, cfgProvider)
	defaultLogger := slog.Default()
	api := api.New(defaultLogger, client, ctrl)

	defaultLogger.Info("Starting app")
	panic(api.Start())
}
