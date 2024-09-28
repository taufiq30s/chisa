package bot

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

var client *redis.Client

// Open redis connection.
// This client will open 10 pool connections
func OpenRedis() {
	ctx := context.Background()
	utils.InfoLog.Println("Opening Redis Connection")
	defer utils.InfoLog.Println("Redis client connected")

	connectionUrl, err := utils.GetEnv("REDIS_URL")
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to get env: %s\n", err)
	}

	opt, err := redis.ParseURL(connectionUrl)
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to parsing connection string. %s\n", err)
	}
	client = redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		utils.ErrorLog.Fatalf("Failed to connect redis. %s\n", err)
	}
}

// Get Redis
func GetRedis() *redis.Client {
	return client
}

// Close Redis Connection
func CloseRedis() {
	utils.InfoLog.Println("Closing Redis Connection")
	defer utils.InfoLog.Println("Redis client closed")

	err := client.Close()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to close redis connection. %s", err)
	}
}
