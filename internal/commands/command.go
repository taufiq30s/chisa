package commands

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/music"
)

var commands []*discordgo.ApplicationCommand
var commandHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)

func init() {
	commands = append(commands,
		music.GetCommands()...,
	)
	commandHandlers = map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate){
		"music": music.GetCommandHandlers(),
	}
}

func registerHandler(chisa *bot.Bot) {
	log.Println("Registering Command Handles")
	chisa.Session.AddHandler(func(c *discordgo.Session, interaction *discordgo.InteractionCreate) {
		handle, ok := commandHandlers[interaction.ApplicationCommandData().Name]
		if ok {
			handle(chisa, interaction)
		}
	})
}

func Unregister(chisa *bot.Bot, guildId *string, registeredCommands []*discordgo.ApplicationCommand) {
	log.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := chisa.Session.ApplicationCommandDelete(chisa.Session.State.User.ID, *guildId, command.ID)
		if err != nil {
			log.Fatalf("Failed to delete '%v' command: %v", command.Name, err)
			break
		}
	}
}

func Register(chisa *bot.Bot, guildId *string) {
	log.Println("Registering Commands")
	registerCommands := make([]*discordgo.ApplicationCommand, len(commands))
	isFailed := false
	for i, command := range commands {
		cmd, err := chisa.Session.ApplicationCommandCreate(chisa.Session.State.User.ID, *guildId, command)
		if err != nil {
			log.Fatalf("Failed to create '%v' command: %v", command.Name, err)
			log.Fatal("Executing Rollback")
			isFailed = true
			break
		}
		registerCommands[i] = cmd
	}

	if isFailed {
		Unregister(chisa, guildId, registerCommands)
		return
	}
	registerHandler(chisa)
	defer log.Println("Registering Commands Successfully")
}
