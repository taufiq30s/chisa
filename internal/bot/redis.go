package bot

import (
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

type Redis struct {
	Client *redis.Client
}

func OpenRedis() Redis {
	log.Println("Opening Redis Connection")
	defer log.Println("Redis client connected")

	connectionUrl, err := utils.GetEnv("REDIS_URL")
	if err != nil {
		log.Fatal(err)
	}

	opt, err := redis.ParseURL(connectionUrl)
	if err != nil {
		log.Fatalf("Failed to parsing connection string. %s", err)
	}
	return Redis{
		Client: redis.NewClient(opt),
	}
}

func (r *Redis) CloseRedis() {
	log.Println("Closing Redis Connection")
	defer log.Println("Redis client closed")

	err := r.Client.Close()
	if err != nil {
		log.Fatalf("Failed to close redis connection. %s", err)
	}
}
