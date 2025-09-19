package main

import (
	"log/slog"

	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/api"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/config"
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/controller/notification"
	rlredis "github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/ratelimiter/redis"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	rateLimiter := rlredis.New(client)

	configs, err := config.LoadFromJson("./configs/limits.json")
	if err != nil {
		panic(err)
	}
	cfgProvider := config.NewRLConfigProvider(configs)
	ctrl := notification.NewController(rateLimiter, cfgProvider)
	defaultLogger := slog.Default()
	api := api.New(defaultLogger, client, ctrl)

	defaultLogger.Info("Starting app")
	api.Start()
}
