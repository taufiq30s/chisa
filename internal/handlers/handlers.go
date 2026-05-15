package handlers

import (
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/events"
	"github.com/taufiq30s/chisa/internal/quiz"
)

type componentFunction func(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{})

// Registry holds all Discord command and event handler registrations for the bot.
type Registry struct {
	bot                     *bot.Bot
	commands                []*discordgo.ApplicationCommand
	eventHandlers           []func(chisa *bot.Bot) interface{}
	commandHandlers         map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	commandAutofillHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	buttonHandlers          map[string]componentFunction
	selectHandlers          map[string]componentFunction
}

// NewRegistry creates a Registry bound to the given Bot, wiring all command
// definitions and handler functions. Call Register to push them to Discord.
func NewRegistry(b *bot.Bot) *Registry {
	r := &Registry{bot: b}

	r.commands = slices.Concat(r.commands,
		musicCommands,
		VerificationCommands,
		currencyCommands,
		quizCommands,
	)

	r.commandHandlers = map[string]func(*bot.Bot, *discordgo.InteractionCreate){
		"music":             musicCommandHandler,
		"verify":            VerificationCommandHandlers,
		"currency":          currencyCommandHandler,
		"interactive-quiz":  quizCommandHandler,
	}

	r.commandAutofillHandlers = map[string]func(*bot.Bot, *discordgo.InteractionCreate){
		"currency": currencyCommandOptions,
	}

	r.buttonHandlers = mergeMap(
		ScamButtonResponseHandler,
		VerificationButtonResponseHandler,
		map[string]componentFunction{
			"search-next": func(s *discordgo.Session, i *discordgo.InteractionCreate, _ ...interface{}) {
				b.Music.SearchNextPage(s, i)
			},
			"search-previous": func(s *discordgo.Session, i *discordgo.InteractionCreate, _ ...interface{}) {
				b.Music.SearchPreviousPage(s, i)
			},
		},
	)

	r.selectHandlers = mergeMap(
		map[string]componentFunction{
			"search-select": func(s *discordgo.Session, i *discordgo.InteractionCreate, _ ...interface{}) {
				data := i.MessageComponentData()
				b.Music.SearchSelect(s, i, data.Values[0])
			},
		},
	)

	r.eventHandlers = []func(*bot.Bot) interface{}{
		events.MessageCreate,
		events.OnVoiceServerUpdate,
		events.OnVoiceStateUpdate,
		func(b *bot.Bot) interface{} { return quiz.OnVoiceStateUpdate(b.Quiz) },
	}

	return r
}

// mergeMap merges multiple componentFunction maps into one.
func mergeMap(maps ...map[string]componentFunction) map[string]componentFunction {
	merged := make(map[string]componentFunction)
	for _, m := range maps {
		for key, value := range m {
			merged[key] = value
		}
	}
	return merged
}
