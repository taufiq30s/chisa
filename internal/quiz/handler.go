package quiz

import (
	"github.com/bwmarrin/discordgo"
)

// OnVoiceStateUpdate returns a raw discordgo handler function that forwards
// VOICE_STATE_UPDATE events to the active quiz session.
// It runs alongside the music voice handler; the two don't interfere because
// the music handler only acts on the bot's own voice state.
//
// The caller (handlers.NewRegistry) wraps this in a func(*bot.Bot) interface{}
// closure so the quiz package never needs to import the bot package.
func OnVoiceStateUpdate(quiz QuizService) func(*discordgo.Session, *discordgo.VoiceStateUpdate) {
	return func(s *discordgo.Session, e *discordgo.VoiceStateUpdate) {
		// Skip the bot's own voice state — handled by the music voice handler.
		if e.UserID == s.State.User.ID {
			return
		}
		if quiz == nil {
			return
		}
		quiz.HandleVoiceStateUpdate(s, e)
	}
}
