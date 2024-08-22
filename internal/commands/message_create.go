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

	if len(m.Content) < 10 {
		return
	}

	go func() {
		client := bot.GetRedis()
		isScam := moderation.CheckScam(client, s, m)
		if isScam > 0 {
			moderation.HandleScamMessage(s, m, isScam)
		}
	}()
}
