package cronjob

import (
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/utils"
)

func updateScamDataset() {
	utils.InfoLog.Println("Updating scam datasets")
	redis := bot.GetRedis()
	err := moderation.UpdateDataset(redis)
	if err != nil {
		utils.ErrorLog.Printf("Failed to update dataset: %v\n", err)
		return
	}
	utils.InfoLog.Println("Updated successfully")
}
