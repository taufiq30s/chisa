package handlers

import (
	"maps"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/events"
	"github.com/taufiq30s/chisa/internal/music"
	"github.com/taufiq30s/chisa/internal/quiz"
)

// componentFunction is the signature shared by every button/select handler.
// params carries extra data a future component may need; today all handlers
// read what they need from the interaction itself.
type componentFunction func(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{})

// Registry holds all Discord command and event handler registrations for the bot.
type Registry struct {
	bot                     *bot.Bot
	commands                []*discordgo.ApplicationCommand
	eventHandlers           []func(chisa *bot.Bot) interface{}
	commandHandlers         map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	commandAutofillHandlers map[string]func(chisa *bot.Bot, i *discordgo.InteractionCreate)
	// buttonHandlers and selectHandlers are keyed by CustomID *prefix*, not by
	// the full CustomID. Discord only allows 100 characters and a single static
	// ID would clash between two users' search results, so the music buttons
	// embed the interaction ID: "search-next-1234...". lookupComponentHandler
	// does the prefix match.
	buttonHandlers map[string]componentFunction
	selectHandlers map[string]componentFunction
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
		"music":            musicCommandHandler,
		"verify":           VerificationCommandHandlers,
		"currency":         currencyCommandHandler,
		"interactive-quiz": quizCommandHandler,
	}

	r.commandAutofillHandlers = map[string]func(*bot.Bot, *discordgo.InteractionCreate){
		"currency": currencyCommandOptions,
	}

	// Music was initialised lazily and may be nil (missing YOUTUBE_API_KEY, for
	// example), so every music component checks it. Ack the interaction even
	// when it is nil, otherwise Discord tells the user "This interaction failed".
	withMusic := func(fn func(music.MusicService, *discordgo.Session, *discordgo.InteractionCreate)) componentFunction {
		return func(s *discordgo.Session, i *discordgo.InteractionCreate, _ ...interface{}) {
			if b.Music == nil {
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredMessageUpdate,
				})
				return
			}
			fn(b.Music, s, i)
		}
	}

	r.buttonHandlers = make(map[string]componentFunction)
	maps.Copy(r.buttonHandlers, ScamButtonResponseHandler)
	maps.Copy(r.buttonHandlers, VerificationButtonResponseHandler)
	maps.Copy(r.buttonHandlers, map[string]componentFunction{
		"search-next": withMusic(func(m music.MusicService, s *discordgo.Session, i *discordgo.InteractionCreate) {
			m.SearchNextPage(s, i)
		}),
		"search-previous": withMusic(func(m music.MusicService, s *discordgo.Session, i *discordgo.InteractionCreate) {
			m.SearchPreviousPage(s, i)
		}),
		"queue-next": withMusic(func(m music.MusicService, s *discordgo.Session, i *discordgo.InteractionCreate) {
			m.ListQueueNextPage(s, i)
		}),
		"queue-previous": withMusic(func(m music.MusicService, s *discordgo.Session, i *discordgo.InteractionCreate) {
			m.ListQueuePreviousPage(s, i)
		}),
	})

	r.selectHandlers = map[string]componentFunction{
		"search-select": withMusic(func(m music.MusicService, s *discordgo.Session, i *discordgo.InteractionCreate) {
			values := i.MessageComponentData().Values
			if len(values) == 0 {
				// Discord always sends one value for a select menu, but
				// indexing blindly would panic and kill the bot.
				return
			}
			m.SearchSelect(s, i, values[0])
		}),
	}

	r.eventHandlers = []func(*bot.Bot) interface{}{
		events.MessageCreate,
		events.OnVoiceStateUpdate,
		func(b *bot.Bot) interface{} { return quiz.OnVoiceStateUpdate(b.Quiz) },
	}

	return r
}
