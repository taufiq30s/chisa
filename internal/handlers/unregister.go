package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

func (r *Registry) unregisterCommands(guildId string, registeredCommands []*discordgo.ApplicationCommand) {
	utils.InfoLog.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := r.bot.Session.ApplicationCommandDelete(r.bot.Session.State.User.ID, guildId, command.ID)
		if err != nil {
			utils.ErrorLog.Printf("Failed to delete '%v' command: %v\n", command.Name, err)
			break
		}
	}
}

func (r *Registry) getCommands(guildId string) []*discordgo.ApplicationCommand {
	commands, err := r.bot.Session.ApplicationCommands(r.bot.Session.State.User.ID, guildId)
	if err != nil {
		utils.ErrorLog.Println(err)
	}
	return commands
}
