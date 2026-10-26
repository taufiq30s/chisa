package events

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
)

// OnVoiceStateUpdate watches the bot's own voice state so the music player can
// clean up when the bot is moved or kicked out of a channel.
//
// Discord does not tell us "you were disconnected"; it sends a voice state
// update with an empty ChannelID. Leaving playback running in that state would
// keep the player goroutine writing into a dead voice connection, so tear
// everything down instead.
func OnVoiceStateUpdate(chisa *bot.Bot) interface{} {
	return func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		// Every user in the guild produces these events, so ignore everyone
		// but the bot itself.
		if s.State.User == nil || e.UserID != s.State.User.ID {
			return
		}
		// chisa.Music is nil when music.New failed during startup.
		if e.ChannelID == "" && chisa.Music != nil {
			chisa.Music.HandleVoiceLeave()
		}
	}
}
