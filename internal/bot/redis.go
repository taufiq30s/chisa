package bot

import (
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

var client *redis.Client

// Open redis connection.
// This client will open 10 pool connections
func OpenRedis() {
	log.Println("Opening Redis Connection")
	defer log.Println("Redis client connected")

	connectionUrl, err := utils.GetEnv("REDIS_URL")
	if err != nil {
		log.Fatalf("Failed to get env: %s", err)
	}

	opt, err := redis.ParseURL(connectionUrl)
	if err != nil {
		log.Fatalf("Failed to parsing connection string. %s", err)
	}
	client = redis.NewClient(opt)
}

// Get Redis
func GetRedis() *redis.Client {
	return client
}

// Close Redis Connection
func CloseRedis() {
	log.Println("Closing Redis Connection")
	defer log.Println("Redis client closed")

	err := client.Close()
	if err != nil {
		log.Fatalf("Failed to close redis connection. %s", err)
	}
}
