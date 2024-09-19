package handlers

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
)

func registerEvents(chisa *bot.Bot) {
	log.Println("Registering Events")
	for _, handler := range eventHandlers {
		chisa.Session.AddHandler(handler)
	}
}

func registerButtonHandlers(id string) (func(chisa *bot.Bot, i *discordgo.InteractionCreate), bool) {
	log.Println("Registering Button Handlers")
	for key := range buttonHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return buttonHandlers[key], true
		}
	}
	return nil, false
}

func registerCommandHandlers(chisa *bot.Bot) {
	log.Println("Registering Command Handlers")
	chisa.Session.AddHandler(func(c *discordgo.Session, interaction *discordgo.InteractionCreate) {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			if handle, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
				handle(chisa, interaction)
			}
		case discordgo.InteractionMessageComponent:
			switch interaction.MessageComponentData().ComponentType {
			case discordgo.ButtonComponent:
				if handle, ok := registerButtonHandlers(interaction.MessageComponentData().CustomID); ok {
					handle(chisa, interaction)
				}
			}
		}
	})
}

func registerCommand(chisa *bot.Bot, guildId string) {
	log.Println("Registering Commands")
	registerCommands := make([]*discordgo.ApplicationCommand, len(commands))
	isFailed := false
	for i, command := range commands {
		cmd, err := chisa.Session.ApplicationCommandCreate(chisa.Session.State.User.ID, guildId, command)
		if err != nil {
			log.Fatalf("Failed to create '%v' command: %v", command.Name, err)
			log.Fatal("Executing Rollback")
			isFailed = true
			break
		}
		registerCommands[i] = cmd
	}

	if isFailed {
		unregisterCommands(chisa, guildId, registerCommands)
		return
	}
}
