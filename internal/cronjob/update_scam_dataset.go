package cronjob

import (
	"log"

	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/moderation"
)

func updateScamDataset() {
	log.Println("Updating scam datasets")
	redis := bot.OpenRedis()
	err := moderation.UpdateDataset(redis.Client)
	if err != nil {
		log.Fatalf("Failed to update dataset: %v", err)
		return
	}
	log.Println("Updated successfully")
	defer redis.CloseRedis()
}
