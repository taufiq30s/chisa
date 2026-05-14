package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

const (
	// rateLimitWindow is the sliding window for per-user command rate limiting.
	rateLimitWindow = 10 * time.Second
	// rateLimitMax is the maximum number of commands a user may issue per window.
	rateLimitMax = 5
)

// checkRateLimit returns true if the interaction should be allowed to proceed.
// It increments a Redis counter keyed by user+command and rejects if over limit.
func checkRateLimit(rdb *redis.Client, s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if rdb == nil {
		return true // no Redis — allow all
	}

	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}
	if userID == "" {
		return true
	}

	cmdName := i.ApplicationCommandData().Name
	key := fmt.Sprintf("rl:%s:%s", userID, cmdName)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		utils.WarningLog.Printf("Rate limit Redis error for user %s: %v\n", userID, err)
		return true // fail open
	}
	if count == 1 {
		// Set expiry on first increment
		rdb.Expire(ctx, key, rateLimitWindow)
	}

	if count > rateLimitMax {
		ttl, _ := rdb.TTL(ctx, key).Result()
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     "Chisa",
			Title:       "Too many requests",
			Description: fmt.Sprintf("You're sending commands too fast. Try again in %.0f seconds.", ttl.Seconds()),
		}).Execute()
		return false
	}
	return true
}
