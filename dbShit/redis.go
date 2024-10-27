package dbShit

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
)

var ctx = context.Background()

// InitializeRedisClient creates a new Redis client
func InitializeRedisClient() *redis.Client {
	println(os.Getenv("REDIS_HOST"))
	log.Println("Connecting to Redis " + os.Getenv("REDIS_HOST") + " with password " + os.Getenv("REDIS_PASSWORD"))
	dbNum, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNum, // Default DB
	})

	return rdb
}
