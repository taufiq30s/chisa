package handlers

import (
	"log"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/internal/music"
)

var (
	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	buttonHandlers  map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
)

func mergeMap(maps ...map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)) map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate) {
	merged := make(map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate))

	for _, m := range maps {
		for key, value := range m {
			merged[key] = value
		}
	}

	return merged
}

// Collect all handlers
func init() {
	commands = slices.Concat(commands,
		music.GetCommands(),
		moderation.VerificationCommands,
	)
	commandHandlers = map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate){
		"music":  music.GetCommandHandlers(),
		"verify": moderation.VerificationCommandHandlers,
	}
	buttonHandlers = mergeMap(
		moderation.GetScamButtonHandlers(),
		moderation.VerificationResponseButtonHandle(),
	)
}

func registerButtonHandlers(id string) (func(chisa *bot.Bot, i *discordgo.InteractionCreate), bool) {
	for key := range buttonHandlers {
		if len(id) >= len(key) && id[:len(key)] == key {
			return buttonHandlers[key], true
		}
	}
	return nil, false
}

func registerHandlers(chisa *bot.Bot) {
	log.Println("Registering Handlers")
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
	// Add Handle when message create
	chisa.Session.AddHandler(messageCreate)
}

func unregisterCommands(chisa *bot.Bot, guildId *string, registeredCommands []*discordgo.ApplicationCommand) {
	log.Println("Unregister Commands")
	for _, command := range registeredCommands {
		err := chisa.Session.ApplicationCommandDelete(chisa.Session.State.User.ID, *guildId, command.ID)
		if err != nil {
			log.Fatalf("Failed to delete '%v' command: %v", command.Name, err)
			break
		}
	}
}

func RegisterCommands(chisa *bot.Bot, guildId *string) {
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
		unregisterCommands(chisa, guildId, registerCommands)
		return
	}
	registerHandlers(chisa)
	defer log.Println("Registering Commands Successfully")
}

func UnregisterCommands(chisa *bot.Bot, guildId *string) {
	log.Println("Unregistering Commands")
	commands, err := chisa.Session.ApplicationCommands(chisa.Session.State.User.ID, *guildId)
	if err != nil {
		log.Fatal(err)
	}
	unregisterCommands(chisa, guildId, commands)
	defer log.Println("Unregistering Commands Successfully")
}
