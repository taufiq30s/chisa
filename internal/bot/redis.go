package bot

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

// Open Redis connection.
// This client will open connections based on PoolSize in Config.
func (chisa *Bot) OpenRedis() {
	ctx := context.Background()
	utils.InfoLog.Println("Opening Redis Connection")
	defer utils.InfoLog.Println("Redis client connected")

	opt, err := redis.ParseURL(chisa.Config.RedisURL)
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to parsing Redis connection string: %s\n", err)
	}
	opt.PoolSize = chisa.Config.PoolSize
	chisa.Redis = redis.NewClient(opt)
	if err := chisa.Redis.Ping(ctx).Err(); err != nil {
		utils.ErrorLog.Fatalf("Failed to connect Redis: %s\n", err)
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
