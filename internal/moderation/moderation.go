package moderation

import (
	"log"

	"github.com/taufiq30s/chisa/utils"
)

var (
	featureName = "Chisa Moderated System"
)

func getModeratorChannelId() string {
	logChannel, err := utils.GetEnv("CHISA_MOD_CHANNEL_ID")
	if err != nil {
		log.Println(err)
	}
	return logChannel
}
