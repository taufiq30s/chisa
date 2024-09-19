package handlers

import (
	"log"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/events"
	"github.com/taufiq30s/chisa/internal/moderation"
	"github.com/taufiq30s/chisa/internal/music"
)

var (
	commands        []*discordgo.ApplicationCommand
	eventHandlers   []interface{}
	commandHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	buttonHandlers  map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
)

// Merge map of command interactions
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
	eventHandlers = []interface{}{
		events.MessageCreate(),
	}
}

func Register(chisa *bot.Bot, guildId string) {
	log.Println("Registering Handler")
	registerCommand(chisa, guildId)
	registerCommandHandlers(chisa)
	registerEvents(chisa)
	defer log.Println("Registering Handlers Successfully")
}

func Unregister(chisa *bot.Bot, guildId string) {
	log.Println("Unregistering Handler")
	commands = getCommands(chisa, guildId)
	unregisterCommands(chisa, guildId, commands)
	defer log.Println("Unregistering Handler Successfully")
}
