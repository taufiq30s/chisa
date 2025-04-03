package handlers

import (
	"slices"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/events"
	"github.com/taufiq30s/chisa/utils"
)

func Register(wg *sync.WaitGroup, chisa *bot.Bot, guildId string) {
	defer wg.Done()
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

type componentFunction func(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{})

var (
	commands                []*discordgo.ApplicationCommand
	eventHandlers           []func(chisa *bot.Bot) interface{}
	commandHandlers         map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	commandAutofillHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	buttonHandlers          map[string]componentFunction
	selectHandlers          map[string]componentFunction
)

// Merge map of command interactions
func mergeMap(maps ...map[string]componentFunction) map[string]componentFunction {
	merged := make(map[string]componentFunction)
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
		currencyCommands,
	)
	commandHandlers = map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate){
		"music":    musicCommandHandler,
		"verify":   VerificationCommandHandlers,
		"currency": currencyCommandHandler,
	}
	commandAutofillHandlers = map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate){
		"currency": currencyCommandOptions,
	}
	buttonHandlers = mergeMap(
		ScamButtonResponseHandler,
		VerificationButtonResponseHandler,
		SearchButtonHandler,
	)
	selectHandlers = mergeMap(
		SearchSelectHandler,
	)
	eventHandlers = []func(chisa *bot.Bot) interface{}{
		events.MessageCreate,
		events.OnVoiceServerUpdate,
		events.OnVoiceStateUpdate,
	}
}
