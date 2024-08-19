package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/moderation"
)

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	go func() {
		// Open Redis connection
		client := bot.OpenRedis()

		isScam := moderation.CheckScam(&client, s, m)
		if isScam {
			return
		}
		defer client.CloseRedis()
	}()
}
