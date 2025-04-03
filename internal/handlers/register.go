package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/utils"
)

func registerEvents(chisa *bot.Bot) {
	fmt.Println("Registering Events")
	utils.InfoLog.Println("Registering Events")
	for _, handler := range eventHandlers {
		fn := handler(chisa)
		chisa.Session.AddHandler(fn)
	}
}

func registerButtonHandlers(id string) (componentFunction, bool) {
	for key := range buttonHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return buttonHandlers[key], true
		}
	}
	return nil, false
}

func registerSelectHandlers(id string) (componentFunction, bool) {
	for key := range selectHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return selectHandlers[key], true
		}
	}
	return nil, false
}

func registerCommandHandlers(chisa *bot.Bot) {
	fmt.Println("Registering Command Handlers")
	utils.InfoLog.Println("Registering Command Handlers")
	chisa.Session.AddHandler(func(c *discordgo.Session, interaction *discordgo.InteractionCreate) {
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			if handle, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
				handle(chisa, interaction)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			if handle, ok := commandAutofillHandlers[interaction.ApplicationCommandData().Name]; ok {
				handle(chisa, interaction)
			}
		case discordgo.InteractionMessageComponent:
			switch interaction.MessageComponentData().ComponentType {
			case discordgo.ButtonComponent:
				if handle, ok := registerButtonHandlers(interaction.MessageComponentData().CustomID); ok {
					if strings.HasPrefix(interaction.MessageComponentData().CustomID, "search") {
						handle(chisa.Session, interaction, chisa.Music)
						return
					}
					handle(chisa.Session, interaction)
				}
			case discordgo.SelectMenuComponent:
				if handle, ok := registerSelectHandlers(interaction.MessageComponentData().CustomID); ok {
					if strings.HasPrefix(interaction.MessageComponentData().CustomID, "search") {
						handle(chisa.Session, interaction, chisa.Music)
						return
					}
					handle(chisa.Session, interaction)
				}
			}
		}
	})
}

func registerCommand(chisa *bot.Bot, guildId string) {
	utils.InfoLog.Println("Registering Commands")
	registerCommands := make([]*discordgo.ApplicationCommand, len(commands))
	isFailed := false
	for i, command := range commands {
		bot, err := chisa.Session.ApplicationCommandCreate(chisa.Session.State.User.ID, guildId, command)
		if err != nil {
			utils.ErrorLog.Printf("Failed to create '%v' command: %v\n", command.Name, err)
			isFailed = true
			break
		}
		registerCommands[i] = bot
	}

	if isFailed {
		utils.InfoLog.Println("Executing Rollback")
		unregisterCommands(chisa, guildId, registerCommands)
		return
	}
}
