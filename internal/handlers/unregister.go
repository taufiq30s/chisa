package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/utils"
)

func unregisterCommands(chisa *bot.Bot, guildId string, registeredCommands []*discordgo.ApplicationCommand) {
	utils.InfoLog.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := chisa.Session.ApplicationCommandDelete(chisa.Session.State.User.ID, guildId, command.ID)
		if err != nil {
			utils.ErrorLog.Printf("Failed to delete '%v' command: %v\n", command.Name, err)
			break
		}
	}
}

func getCommands(chisa *bot.Bot, guildId string) []*discordgo.ApplicationCommand {
	commands, err := chisa.Session.ApplicationCommands(chisa.Session.State.User.ID, guildId)
	if err != nil {
		utils.ErrorLog.Println(err)
	}
	return commands
}
