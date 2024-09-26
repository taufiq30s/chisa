package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/spotify"
	"github.com/taufiq30s/chisa/utils"
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
		utils.ErrorLog.Fatalf("Failed to created discord session: %s\n", err)
	}

	utils.InfoLog.Println("Connecting to bot")
	err = bot.Session.Open()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to open connection: %s\n", err)
	}
	utils.InfoLog.Println("Bot connection open")
}

func (bot *Bot) InitializeSpotifyClient() {
	utils.InfoLog.Println("Connecting to spotify")
	var err error = nil
	bot.SpotifySession, err = spotify.New()
	if err != nil {
		utils.ErrorLog.Printf("Failed to open spotify client session: %s\n", err)
	}
	utils.InfoLog.Println("Spotify client connected")
}

func (bot *Bot) Disconnect() {
	err := bot.Session.Close()
	if err != nil {
		utils.ErrorLog.Fatalf("Failed to close connection: %s\n", err)
	}
	utils.InfoLog.Println("Bot connection closed")
}
