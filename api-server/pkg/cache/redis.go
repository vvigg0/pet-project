package rds

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func Init() {
	Client = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	pong, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Не удалось подключиться к Redis: %v", err)
	}
	log.Printf("Redis подключён успешно: %s", pong)
}
