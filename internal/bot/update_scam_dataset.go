package bot

import (
	"time"

	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/utils"
)

const (
	maxRetryAttempts = 3
	retryBaseDelay   = 5 * time.Second
)

func (chisa *Bot) updateScamDataset() {
	utils.InfoLog.Println("Updating scam datasets")
	var lastErr error
	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		lastErr = moderation.UpdateDataset(chisa.Redis)
		if lastErr == nil {
			utils.InfoLog.Println("Updated successfully")
			return
		}
		utils.ErrorLog.Printf("Failed to update dataset (attempt %d/%d): %v\n", attempt, maxRetryAttempts, lastErr)
		if attempt < maxRetryAttempts {
			backoff := retryBaseDelay * time.Duration(attempt)
			utils.InfoLog.Printf("Retrying in %s...\n", backoff)
			time.Sleep(backoff)
		}
	}
	utils.ErrorLog.Printf("Gave up updating scam dataset after %d attempts: %v\n", maxRetryAttempts, lastErr)
}
