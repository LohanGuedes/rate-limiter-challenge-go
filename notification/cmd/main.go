package main

import (
	"github.com/LohanGuedes/modak-rate-limit-challenge/notification/internal/api"
	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		// TODO: Use an environment variable, even better if using docker-compose!
		Addr: "localhost:6379",
	})

	defer client.Close()

	api := api.New(client)
	api.Start()
}
