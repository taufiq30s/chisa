package moderation

import (
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/config"
)

var (
	featureName = "Chisa Moderated System"
	cfg         *config.Config
	redisClient *redis.Client
)

// SetConfig initialises the moderation package with the application config.
// Must be called before any moderation functions are used.
func SetConfig(c *config.Config) {
	cfg = c
}

// SetRedis gives the moderation package the shared Redis client, used to
// persist verification decisions that outlive the process.
func SetRedis(client *redis.Client) {
	redisClient = client
}

func getVerifiedRoleId() string {
	return cfg.VerifiedRoleID
}

func getVerificationChannelId() string {
	return cfg.VerifChannelID
}

func getModeratorChannelId() string {
	return cfg.ModChannelID
}

func getLogChannel() string {
	return cfg.LogChannelID
}

func getHoneypotChannelID() string {
	return cfg.HoneypotChannelID
}
