package music

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/model"
)

type trackState struct {
	track       model.Track
	interaction *discordgo.InteractionCreate
	addedBy     *string
	channelId   *string
	// ownsDefer means this track's interaction still owes Discord a first
	// reply, so its "Now Playing" card may edit the deferred response
	// (SendDefer). Only a track queued while nothing was playing qualifies:
	// the player loop picks it up before anything answers its /music play.
	// Every later track spends its defer on its own "Added to Queue" card, so
	// reusing it for "Now Playing" would edit an already-answered interaction
	// (or an expired token) and the user would see nothing new.
	ownsDefer bool
}

// enqueue is the single place a track enters the queue. Callers decide whether
// the confirmation card goes through an interaction (state.interaction != nil)
// or straight to the channel (state.interaction == nil).
func (m *MusicBot) enqueue(state trackState) {
	m.queueMu.Lock()
	isMusicPlaying := m.isPlaying
	// A track is about to be queued, so a pending idle disconnect must not fire
	// and tear down the connection we are about to use.
	m.cancelIdleDisconnectLocked()
	// Decided here, while isPlaying is still authoritative: the player loop is
	// what flips it, so a track that arrives before playback starts is the one
	// the loop answers through its interaction.
	state.ownsDefer = !isMusicPlaying && owesFirstReply(state.interaction)
	m.queue = append(m.queue, state)
	m.queueMu.Unlock()

	if isMusicPlaying {
		m.generateMusicInformation(state, "Added to Queue")
	}

	m.startPlayerLoopIfNeeded()
}

// owesFirstReply reports whether the handler's deferred acknowledgement is
// still unanswered. Slash commands defer up front; message components (the
// search picker) were answered with a DeferredMessageUpdate whose target
// message is deleted before the track is queued, so editing it would fail.
func owesFirstReply(i *discordgo.InteractionCreate) bool {
	return i != nil && i.Type == discordgo.InteractionApplicationCommand
}

func (m *MusicBot) addToQueue(track model.Track, i *discordgo.InteractionCreate) {
	m.enqueue(trackState{
		track:       track,
		interaction: i,
		addedBy:     &i.Member.User.Username,
		channelId:   &i.ChannelID,
	})
}

// playlistStates builds one trackState per track for a playlist import.
//
// ownsDefer stays false for every state on purpose. The summary card spends this
// command's deferred reply, so nothing else may SendDefer on the same
// interaction; the first track's "Now Playing" is posted to the channel
// instead, exactly like any later track. Split out from addPlaylistToQueue so
// that rule is testable without a Discord session.
func playlistStates(i *discordgo.InteractionCreate, tracks []model.Track) []trackState {
	addedBy := i.Member.User.Username
	channelID := i.ChannelID
	states := make([]trackState, 0, len(tracks))
	for _, track := range tracks {
		states = append(states, trackState{
			track:       track,
			interaction: i,
			addedBy:     &addedBy,
			channelId:   &channelID,
		})
	}
	return states
}

// addPlaylistToQueue appends every track of a playlist in one shot.
//
// It does NOT loop over enqueue: enqueue sends one card per track and restarts
// the idle timer each time, so a 50-track playlist would spam 50 messages. One
// lock section, one summary card, one player start.
func (m *MusicBot) addPlaylistToQueue(i *discordgo.InteractionCreate, tracks []model.Track) {
	states := playlistStates(i, tracks)

	m.queueMu.Lock()
	// Tracks are arriving, so a pending idle disconnect must not fire and tear
	// down the connection they need.
	m.cancelIdleDisconnectLocked()
	m.queue = append(m.queue, states...)
	m.queueMu.Unlock()

	m.generatePlaylistInformation(states[0], tracks)
	m.startPlayerLoopIfNeeded()
}

// dequeueLocked pops the head of the queue. Caller MUST hold queueMu.
func (m *MusicBot) dequeueLocked() (*trackState, bool) {
	if len(m.queue) == 0 {
		return nil, false
	}
	trackState := m.queue[0]
	m.queue = m.queue[1:]
	return &trackState, true
}

// ClearQueue drops every pending track. It does NOT stop the current track:
// that one has already been dequeued and handed to the player goroutine, so it
// lives in that goroutine's local state, not in m.queue. To stop audio as well,
// call stopPlayback (see MusicBot.Stop and MusicBot.Disconnect).
//
// Safe to call from anywhere - it takes the lock itself. Do not call it while
// already holding queueMu.
func (m *MusicBot) ClearQueue() {
	m.queueMu.Lock()
	m.queue = []trackState{}
	m.queueMu.Unlock()
}
