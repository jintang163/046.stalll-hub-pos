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

func HDel(key string, fields ...string) error {
	return Client.HDel(Ctx, key, fields...).Err()
}

func LPush(key string, values ...interface{}) error {
	return Client.LPush(Ctx, key, values...).Err()
}

func RPush(key string, values ...interface{}) error {
	return Client.RPush(Ctx, key, values...).Err()
}

func LPop(key string) (string, error) {
	return Client.LPop(Ctx, key).Result()
}

func RPop(key string) (string, error) {
	return Client.RPop(Ctx, key).Result()
}

func LLen(key string) (int64, error) {
	return Client.LLen(Ctx, key).Result()
}

func LRange(key string, start, stop int64) ([]string, error) {
	return Client.LRange(Ctx, key, start, stop).Result()
}

func LIndex(key string, index int64) (string, error) {
	return Client.LIndex(Ctx, key, index).Result()
}

func LRem(key string, count int64, value interface{}) error {
	return Client.LRem(Ctx, key, count, value).Err()
}

func ZAdd(key string, score float64, member interface{}) error {
	return Client.ZAdd(Ctx, key, redis.Z{Score: score, Member: member}).Err()
}

func ZRange(key string, start, stop int64) ([]string, error) {
	return Client.ZRange(Ctx, key, start, stop).Result()
}

func ZRem(key string, members ...interface{}) error {
	return Client.ZRem(Ctx, key, members...).Err()
}

func ZRank(key string, member string) (int64, error) {
	return Client.ZRank(Ctx, key, member).Result()
}

func ZCard(key string) (int64, error) {
	return Client.ZCard(Ctx, key).Result()
}

func Incr(key string) (int64, error) {
	return Client.Incr(Ctx, key).Result()
}

func Expire(key string, expiration time.Duration) error {
	return Client.Expire(Ctx, key, expiration).Err()
}

func Publish(channel string, message interface{}) error {
	return Client.Publish(Ctx, channel, message).Err()
}

func Subscribe(channels ...string) *redis.PubSub {
	return Client.Subscribe(Ctx, channels...)
}
