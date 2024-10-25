package contact

import (
	"github.com/redis/go-redis/v9"
)

var redisAddr = "redis:6379"
var redisDB = 0

func GetClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})
}
