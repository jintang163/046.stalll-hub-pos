package redis

import (
	"context"
	"fmt"
	"log"
	"stalll-hub-pos/backend/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func InitRedis() {
	cfg := config.AppConfig.Redis
	Client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	if err := Client.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
}

func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(Ctx, key, value, expiration).Err()
}

func Get(key string) (string, error) {
	return Client.Get(Ctx, key).Result()
}

func Del(key string) error {
	return Client.Del(Ctx, key).Err()
}

func Exists(key string) (int64, error) {
	return Client.Exists(Ctx, key).Result()
}

func HSet(key string, values ...interface{}) error {
	return Client.HSet(Ctx, key, values...).Err()
}

func HGetAll(key string) (map[string]string, error) {
	return Client.HGetAll(Ctx, key).Result()
}
