package bot

import (
	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/utils"
)

func (chisa *Bot) updateScamDataset() {
	utils.InfoLog.Println("Updating scam datasets")
	err := moderation.UpdateDataset(chisa.Redis)
	if err != nil {
		utils.ErrorLog.Printf("Failed to update dataset: %v\n", err)
		return
	}
	utils.InfoLog.Println("Updated successfully")
}
