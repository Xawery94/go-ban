package redis

import (
	"log"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type RedisClient struct {
	*redis.Client
}

var redisClient *RedisClient

func GetRedisClient() (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", //TODO how to insert value from config ??
		Password: "",
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect to Redis")
	}

	return redisClient, nil
}

func (c *RedisClient) BanClient(ip string, ttl time.Duration) {
	c.Set(ip, nil, ttl)
}

func (c *RedisClient) IsUserBanned(ip string) {
	err := c.Get(ip)

	if err != nil {
		log.Printf("Key %v not found %v", ip, err)
	}
}

func (c *RedisClient) RemoveBan(ip string) {
	c.Del(ip)
}
