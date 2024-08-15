package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

type bot struct {
	Session *discordgo.Session
}

func New() bot {
	return bot{}
}

func (bot *bot) Connect(token string) {
	var err error = nil
	bot.Session, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to created discord session: %s", err)
	}

	err = bot.Session.Open()
	if err != nil {
		log.Fatalf("Failed to open connection: %s", err)
	}
	log.Println("Bot connection open")
}

func (bot *bot) Disconnect() {
	err := bot.Session.Close()
	if err != nil {
		log.Fatalf("Failed to close connection: %s", err)
	}
	log.Println("Bot connection closed")
}
