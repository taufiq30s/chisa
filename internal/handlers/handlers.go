package handlers

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/events"
	"github.com/taufiq30s/chisa/utils"
)

func Register(chisa *bot.Bot, guildId string) {
	utils.InfoLog.Println("Registering Handler")
	defer utils.InfoLog.Println("Registering Handlers Successfully")

	registerCommand(chisa, guildId)
	registerCommandHandlers(chisa)
	registerEvents(chisa)
}

func Unregister(chisa *bot.Bot, guildId string) {
	utils.InfoLog.Println("Unregistering Handler")
	defer utils.InfoLog.Println("Unregistering Handler Successfully")

	commands = getCommands(chisa, guildId)
	unregisterCommands(chisa, guildId, commands)
}

var (
	commands        []*discordgo.ApplicationCommand
	eventHandlers   []func(chisa *bot.Bot) interface{}
	commandHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	buttonHandlers  map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)
)

// Merge map of command interactions
func mergeMap(maps ...map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate)) map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	merged := make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))

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
		musicCommands,
		VerificationCommands,
	)
	commandHandlers = map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate){
		"music":  musicCommandHandler,
		"verify": VerificationCommandHandlers,
	}
	buttonHandlers = mergeMap(
		ScamButtonResponseHandler,
		VerificationButtonResponseHandler,
	)
	eventHandlers = []func(chisa *bot.Bot) interface{}{
		events.MessageCreate,
	}
}
