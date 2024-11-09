package events

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/bot"
)

func OnVoiceStateUpdate(chisa *bot.Bot) interface{} {
	return func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		if e.UserID != s.State.User.ID {
			return
		}

		var channelID *snowflake.ID
		if e.ChannelID != "" {
			id := snowflake.MustParse(e.ChannelID)
			channelID = &id
		}
		chisa.Music.Client.OnVoiceStateUpdate(
			context.TODO(),
			snowflake.MustParse(e.GuildID),
			channelID,
			e.SessionID,
		)
		// if event.ChannelID == "" {
		// 	b.Queues.Delete(event.GuildID)
		// }
	}
}
