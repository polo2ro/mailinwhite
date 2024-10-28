package common

import (
	"github.com/redis/go-redis/v9"
)

var redisAddr = "redis:6379"
var redisDbAddresses = 0
var redisDbMessages = 1

func GetAddressesClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDbAddresses,
	})
}

func GetMessagesClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDbMessages,
	})
}

// structure for message storage in redisDbMessages
type MessageData struct {
	Content []byte   `json:"content"`
	From    string   `json:"from"`
	To      []string `json:"to"`
}
