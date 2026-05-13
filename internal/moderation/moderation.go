package moderation

import (
	"github.com/taufiq30s/chisa/internal/config"
)

var (
	featureName    = "Chisa Moderated System"
	verifiedRoleId string
	cfg            *config.Config
)

// SetConfig initialises the moderation package with the application config.
// Must be called before any moderation functions are used.
func SetConfig(c *config.Config) {
	cfg = c
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
