package music

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/music/providers"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

/*
Play Command
*/
func (m *MusicBot) Load(s *discordgo.Session, i *discordgo.InteractionCreate, data string) {
	// Check the voice requirements up front but do NOT join yet. These checks
	// are cheap and failing early means the user learns about a missing
	// permission before browsing search results; the actual join is deferred
	// until a track is committed (see the two ensureVoiceJoined calls below),
	// so a cancelled search never leaves the bot idling in the channel.
	channelID := m.ensureVoiceState(s, i)
	if channelID == "" {
		return
	}
	if err := m.verifyVoicePermissions(s, channelID); err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while joining voice channel",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}

	// Find track and store selected track.
	// 20s, not 2s: a yt-dlp cold search measures ~2.5s and a slow network can
	// triple that. Too short a deadline cancels yt-dlp mid-search and the user
	// sees "context deadline exceeded" instead of results.
	// A URL gets the playlist-sized deadline: extracting a playlist walks every
	// entry, and a direct video URL only ever needs one page, so sharing the
	// longer deadline there costs nothing but a slow failure instead of an early
	// one. Searches keep the short deadline because the picker must feel snappy.
	resolveTimeout := ResolveTimeout
	if providers.IsURL(data) {
		resolveTimeout = PlaylistResolveTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	result, isSearch, err := m.resolver.Resolve(ctx, data)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while resolving track",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}
	if len(result) == 0 {
		description := "Please check your URL"
		if isSearch {
			// A search that keeps nothing is not a failure: yt-dlp found only
			// live streams, playlists, or nothing at all for this query.
			description = "No playable video matched your search. Try different keywords."
		}
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Track Not Found",
			Description: description,
		}).ExecuteDefer()
		return
	}
	if isSearch {
		m.search(s, i, result)
		return
	}

	// A direct URL is committed right here, so join now. The interaction is
	// still deferred, which is why a join failure can reply normally.
	if err := m.ensureVoiceJoined(s, channelID); err != nil {
		utils.ErrorLog.Println("Failed to join voice channel:", err)
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while joining voice channel",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}

	// A non-search resolve that yields several tracks is a playlist: queue all of
	// them behind one join and one summary card instead of picking result[0].
	if len(result) > 1 {
		m.addPlaylistToQueue(i, result)
		return
	}

	m.addToQueue(result[0], i)
}

// Skip stops the current track and moves straight to the next one.
// Note what is NOT done here: isPlaying stays true and the queue is untouched.
// The player goroutine does all of that itself when Streamer.Play returns, which
// is why Skip is just a cancel - it never races the loop for ownership of state.
func (m *MusicBot) Skip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.queueMu.Lock()
	cancel := m.playCtxCancel
	m.queueMu.Unlock()

	if cancel == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Nothing is playing",
			Description: "There is no track to skip.",
		}).ExecuteDefer()
		return
	}
	cancel()

	embed := responses.CreateMessageEmbed(
		s, "Skipped",
		"Skipped the current track.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SendDefer()
}

// HandleVoiceLeave tears everything down without replying to anything. The
// voice state event calls it when Discord removes the bot from the channel:
// there is no interaction to answer, but the player goroutine still has to
// exit before the voice connection is dropped, and the streamer has to be
// forgotten or the next /music play would reuse a dead connection.
//
// Safe to call twice for one leave: stopPlayback returns immediately when no
// player is running, ClearQueue is idempotent and takeStreamer hands the
// streamer to exactly one caller.
func (m *MusicBot) HandleVoiceLeave() {
	// Discord already pulled the bot out, so there is nothing left to time out.
	m.cancelIdleDisconnect()
	m.stopPlayback()
	m.ClearQueue()
	m.disconnectFromVoiceChannel()
}

// Stop halts playback and empties the queue.
// stopPlayback blocks until the player goroutine has actually exited, so a
// /music play arriving right after this cannot spawn a second player loop.
func (m *MusicBot) Stop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.cancelIdleDisconnect()
	m.stopPlayback()
	m.ClearQueue()

	// Stop empties the queue, which is one of the idle triggers: leave after the
	// grace period instead of dropping the connection immediately, so a quick
	// /music play afterwards reuses it.
	m.scheduleIdleDisconnect()

	embed := responses.CreateMessageEmbed(
		s, "Stopped",
		"The music player has been stopped and the queue has been cleared.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(embed).SendDefer()
}

// ClearQueue empties the pending queue but keeps the current track playing.
// If you ever want "clear" to also stop the music, call stopPlayback first -
// resetting the queue alone does not touch the player.
func (m *MusicBot) ClearQueueNow(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.queueMu.Lock()
	pending := len(m.queue)
	m.queue = []trackState{}
	m.queueMu.Unlock()

	embed := responses.CreateMessageEmbed(
		s, "Queue Cleared",
		fmt.Sprintf("Removed %d track(s) from the queue. The current track keeps playing.", pending),
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SendDefer()
}

// Disconnect stops playback, empties the queue and leaves the voice channel.
// The order matters: stop first, then disconnect. Disconnecting while the
// player is mid-track leaves it writing into a dead voice connection.
func (m *MusicBot) Disconnect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.cancelIdleDisconnect()
	m.HandleVoiceLeave()

	embed := responses.CreateMessageEmbed(
		s, "Disconnected",
		"Left the voice channel and cleared the queue.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SendDefer()
}

// Pause freezes the current track. It does not touch the queue or the player
// loop, so Skip/Stop keep working while paused: both of those cancel the play
// context, which unwinds Streamer.Play even with the children SIGSTOPped.
func (m *MusicBot) Pause(s *discordgo.Session, i *discordgo.InteractionCreate) {
	streamer := m.currentStreamer()
	if streamer == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Nothing is playing",
			Description: "There is no active playback to pause.",
		}).ExecuteDefer()
		return
	}
	if err := streamer.Pause(); err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while pausing",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}
	embed := responses.CreateMessageEmbed(s, "Paused", "Playback paused.", m.featureName, responses.SetColor("0bdd47"))
	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SendDefer()
}

// Resume undoes Pause.
func (m *MusicBot) Resume(s *discordgo.Session, i *discordgo.InteractionCreate) {
	streamer := m.currentStreamer()
	if streamer == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Nothing is playing",
			Description: "There is no active playback to resume.",
		}).ExecuteDefer()
		return
	}
	if err := streamer.Resume(); err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while resuming",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}
	embed := responses.CreateMessageEmbed(s, "Resumed", "Playback resumed.", m.featureName, responses.SetColor("0bdd47"))
	responses.InteractionResponse(s, i.Interaction).WithEmbed(embed).SendDefer()
}
