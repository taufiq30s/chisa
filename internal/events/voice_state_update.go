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
		chisa.Music.ForwardVoiceStateUpdate(
			context.Background(),
			snowflake.MustParse(e.GuildID),
			channelID,
			e.SessionID,
		)
		// When the bot is removed from a voice channel, clear the music queue.
		if e.ChannelID == "" {
			chisa.Music.ClearQueue()
		}
	}
}
