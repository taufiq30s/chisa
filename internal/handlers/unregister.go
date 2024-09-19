package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
)

func unregisterCommands(chisa *bot.Bot, guildId string, registeredCommands []*discordgo.ApplicationCommand) {
	log.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := chisa.Session.ApplicationCommandDelete(chisa.Session.State.User.ID, guildId, command.ID)
		if err != nil {
			log.Fatalf("Failed to delete '%v' command: %v", command.Name, err)
			break
		}
	}
}

func getCommands(chisa *bot.Bot, guildId string) []*discordgo.ApplicationCommand {
	commands, err := chisa.Session.ApplicationCommands(chisa.Session.State.User.ID, guildId)
	if err != nil {
		log.Fatal(err)
	}
	return commands
}
