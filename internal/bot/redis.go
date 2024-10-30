package bot

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

// Open Redis connection.
// This client will open 10 pool connections
func (chisa *Bot) OpenRedis() {
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
	chisa.Redis = redis.NewClient(opt)
	if err := chisa.Redis.Ping(ctx).Err(); err != nil {
		utils.ErrorLog.Fatalf("Failed to connect Redis. %s\n", err)
	}
}

// Close Redis Connection
func (chisa *Bot) CloseRedis() {
	utils.InfoLog.Println("Closing Redis Connection")
	defer utils.InfoLog.Println("Redis client closed")

	err := chisa.Redis.Close()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to close Redis connection. %s", err)
	}
}
