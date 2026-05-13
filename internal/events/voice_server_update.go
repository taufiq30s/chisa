package events

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/bot"
)

func OnVoiceServerUpdate(chisa *bot.Bot) interface{} {
	return func(session *discordgo.Session, event *discordgo.VoiceServerUpdate) {
		chisa.Music.ForwardVoiceServerUpdate(
			context.Background(),
			snowflake.MustParse(event.GuildID),
			event.Token,
			event.Endpoint,
		)
	}
}
