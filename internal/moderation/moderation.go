package moderation

import (
	"log"

	"github.com/taufiq30s/chisa/utils"
)

var (
	featureName    = "Chisa Moderated System"
	verifiedRoleId string
)

func getVerifiedRoleId() string {
	verifiedRoleId, err := utils.GetEnv("AKASHIC_VERIFIED_ROLE_ID")
	if err != nil {
		utils.ErrorLog.Println(err)
	}
	return verifiedRoleId
}

func getVerificationChannelId() string {
	logChannel, err := utils.GetEnv("AKASHIC_VERIF_CHANNEL_ID")
	if err != nil {
		log.Println(err)
	}
	return logChannel
}

func getModeratorChannelId() string {
	logChannel, err := utils.GetEnv("CHISA_MOD_CHANNEL_ID")
	if err != nil {
		log.Println(err)
	}
	return logChannel
}

func getLogChannel() string {
	logChannel, err := utils.GetEnv("CHISA_LOG_CHANNEL_ID")
	if err != nil {
		utils.ErrorLog.Println(err)
	}
	return logChannel
}
