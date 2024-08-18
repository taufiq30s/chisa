package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/spotify"
)

type Bot struct {
	Session        *discordgo.Session
	SpotifySession *spotify.Client
}

func New() Bot {
	return Bot{}
}

func (bot *Bot) Connect(token string) {
	var err error = nil
	bot.Session, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to created discord session: %s", err)
	}

	log.Println("Connecting to bot")
	err = bot.Session.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %s", err)
	}
	log.Println("Bot connection open")
}

func (bot *Bot) InitializeSpotifyClient() {
	var err error = nil
	bot.SpotifySession, err = spotify.New()
	if err != nil {
		log.Fatalf("Failed to open spotify client session: %s", err)
	}
	log.Println("Spotify client connected")
}

func (bot *Bot) Disconnect() {
	err := bot.Session.Close()
	if err != nil {
		log.Fatalf("Failed to close connection: %s", err)
	}
	log.Println("Bot connection closed")
}
