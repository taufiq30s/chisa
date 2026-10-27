package music

import (
	"context"
	"fmt"
	"time"

	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

// startPlayerLoopIfNeeded spawns the player goroutine unless one is already
// running. The check and the spawn happen under queueMu, so two concurrent
// /music play interactions can never create two players. Two players would
// both write to voiceConnection.OpusSend and garble/duplicate the audio.
func (m *MusicBot) startPlayerLoopIfNeeded() {
	m.queueMu.Lock()
	done, ok := m.beginPlaybackLocked()
	m.queueMu.Unlock()
	if !ok {
		return
	}

	go m.playerLoop(done)
}

// beginPlaybackLocked claims the player slot and returns the channel the new
// goroutine must close when it exits. Returns ok == false when a player is
// already running. Caller must hold queueMu.
//
// Split out of startPlayerLoopIfNeeded so the claim can be tested without
// spawning a goroutine that would concurrently mutate the same fields.
func (m *MusicBot) beginPlaybackLocked() (done chan struct{}, ok bool) {
	// A player is starting, so anything the idle timer was about to do is
	// obsolete. Cancelling here as well as in enqueue covers the ordering where
	// the queue drains and refills faster than the timer's callback can run.
	m.cancelIdleDisconnectLocked()
	if m.isPlaying {
		return nil, false
	}

	m.isPlaying = true
	m.stopRequested = false
	done = make(chan struct{})
	m.playerDone = done
	return done, true
}

// stopPlayback asks the player to stop and blocks until the goroutine has
// really exited. Never call it while holding queueMu (the loop needs it).
//
// Why playerDone exists: isPlaying used to be cleared by whoever called Stop,
// while playerLoop was still unwinding inside Streamer.Play. A /music play
// arriving in that window saw isPlaying == false and spawned a second player,
// so both loops ended up feeding OpusSend. Now only the player goroutine clears
// isPlaying (finishLocked), and callers wait for playerDone, which is closed by
// the goroutine itself right before it returns.
func (m *MusicBot) stopPlayback() {
	m.queueMu.Lock()
	m.stopRequested = true
	cancel := m.playCtxCancel
	done := m.playerDone
	m.queueMu.Unlock()

	if cancel != nil {
		// Kills yt-dlp/ffmpeg (exec.CommandContext) and unblocks the OpusSend
		// select inside Streamer.Play.
		cancel()
	}
	if done != nil {
		<-done
	}
}

// finishLocked resets playback state. Caller must hold queueMu.
func (m *MusicBot) finishLocked() {
	m.isPlaying = false
	m.playCtxCancel = nil
	m.playerDone = nil
	m.stopRequested = false
	m.currentTrack = nil
}

// scheduleIdleDisconnect starts (or restarts) the countdown to leaving the
// voice channel. Called only when the queue is empty and nothing is playing;
// any new track cancels it through enqueue/startPlayerLoopIfNeeded.
//
// The callback runs on its own goroutine with no context, so it must not hold
// queueMu while touching the voice connection or Discord: disconnectFromVoiceChannel
// takes streamerMu and the channel send does network I/O.
func (m *MusicBot) scheduleIdleDisconnect() {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()
	m.scheduleIdleDisconnectLocked()
}

// scheduleIdleDisconnectLocked is the queueMu-holding variant. Caller must hold
// queueMu.
//
// It deliberately does not inspect the streamer: that would take streamerMu
// while holding queueMu, and the reverse order is used by the disconnect path.
// A timer scheduled while not connected is harmless - both
// disconnectFromVoiceChannel and sendNoActivity no-op in that case.
func (m *MusicBot) scheduleIdleDisconnectLocked() {
	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}
	m.idleToken++
	token := m.idleToken
	m.idleTimer = time.AfterFunc(AutoDisconnectDelay, func() {
		m.onIdleTimeout(token)
	})
}

// cancelIdleDisconnect stops a pending idle disconnect. Never call it while
// holding queueMu.
func (m *MusicBot) cancelIdleDisconnect() {
	m.queueMu.Lock()
	defer m.queueMu.Unlock()
	m.cancelIdleDisconnectLocked()
}

// cancelIdleDisconnectLocked is the queueMu-holding variant. Caller must hold
// queueMu.
func (m *MusicBot) cancelIdleDisconnectLocked() {
	m.idleToken++
	if m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}
}

// onIdleTimeout runs AutoDisconnectDelay after the queue drained. The token
// check is the whole point: Timer.Stop cannot guarantee the callback has not
// already been scheduled, so a stale callback must recognise that it was
// cancelled or replaced and return without touching anything.
func (m *MusicBot) onIdleTimeout(token uint64) {
	m.queueMu.Lock()
	if token != m.idleToken || m.isPlaying || len(m.queue) > 0 || m.stopRequested {
		m.queueMu.Unlock()
		return
	}
	// Claim the timer: bumping the token means a /music play that arrives from
	// here on invalidates this callback and the re-check below notices.
	m.idleToken++
	claim := m.idleToken
	m.idleTimer = nil
	channelID := m.lastChannelId
	m.queueMu.Unlock()

	// Re-check immediately before leaving voice. The gap between the decision
	// and disconnectFromVoiceChannel is non-zero, and a track queued in that
	// window must win over the teardown.
	m.queueMu.Lock()
	stillIdle := claim == m.idleToken && !m.isPlaying && len(m.queue) == 0
	m.queueMu.Unlock()
	if !stillIdle {
		return
	}

	m.disconnectFromVoiceChannel()
	m.sendNoActivity(channelID)
}

// sendNoActivity reports the automatic disconnect. It goes to the channel the
// last track was played in, because the interaction that started playback is
// long gone by the time the timer fires.
func (m *MusicBot) sendNoActivity(channelID string) {
	if channelID == "" {
		return
	}
	embed := responses.CreateMessageEmbed(
		m.session,
		"No Activity",
		fmt.Sprintf("Left the voice channel because nothing played for %s.", formatIdleDelay()),
		m.featureName,
		responses.SetColor("c30010"),
	)
	if _, err := m.session.ChannelMessageSendEmbed(channelID, embed); err != nil {
		utils.ErrorLog.Println("Failed to send no-activity message:", err)
	}
}

// formatIdleDelay renders AutoDisconnectDelay for the user-facing message, so
// editing the constant keeps the text honest instead of saying "10 minutes"
// after someone changes it to 2.
func formatIdleDelay() string {
	d := AutoDisconnectDelay
	if d%time.Minute == 0 {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}
	return d.String()
}

// sendPlaybackError posts a playback failure into the channel the command was
// run in. It cannot use the interaction response: by the time the worker runs,
// that interaction is already acknowledged, so any InteractionRespond call
// would fail with 40060 and the user would see nothing.
func (m *MusicBot) sendPlaybackError(state *trackState, description string) {
	if state.channelId == nil {
		return
	}
	embed := responses.CreateMessageEmbed(
		m.session,
		"Error while playing track",
		description,
		m.featureName,
		responses.SetColor("c30010"),
	)
	if _, err := m.session.ChannelMessageSendEmbed(*state.channelId, embed); err != nil {
		utils.ErrorLog.Println("Failed to send error response:", err)
	}
}

func (m *MusicBot) playerLoop(done chan struct{}) {
	defer close(done)

	for {
		// Everything from the shutdown check down to installing playCtxCancel
		// happens under ONE lock. That makes the "stop requested" check and the
		// decision to play a track atomic, which is what stopPlayback relies on:
		// it can either cancel the context we installed here, or observe
		// stopRequested on our next iteration - never neither.
		m.queueMu.Lock()

		select {
		case <-m.shutdown:
			m.finishLocked()
			m.queueMu.Unlock()
			return
		default:
		}

		if m.stopRequested {
			m.finishLocked()
			m.queueMu.Unlock()
			return
		}

		trackState, ok := m.dequeueLocked()
		if !ok {
			// The queue ran dry: this is the natural-drain trigger for the idle
			// disconnect. Scheduled while still holding queueMu so a concurrent
			// enqueue either cancels it or schedules after us.
			m.scheduleIdleDisconnectLocked()
			m.finishLocked()
			m.queueMu.Unlock()
			return
		}
		// Publish the track for /music list. lastChannelId outlives playback so
		// the idle disconnect has somewhere to post.
		m.currentTrack = trackState
		if trackState.channelId != nil {
			m.lastChannelId = *trackState.channelId
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.playCtxCancel = cancel
		m.queueMu.Unlock()

		// Read the streamer once. Reading the field twice would allow
		// disconnectFromVoiceChannel to nil it between the nil check and Play,
		// panicking this goroutine (which would take the whole bot down).
		streamer := m.currentStreamer()
		if streamer == nil {
			utils.ErrorLog.Println("Streamer is not initialized")
			m.sendPlaybackError(trackState, "Streamer is not initialized")
			m.queueMu.Lock()
			m.finishLocked()
			m.queueMu.Unlock()
			return
		}

		m.generateMusicInformation(*trackState, "Now Playing")
		if err := streamer.Play(ctx, trackState.track); err != nil {
			utils.ErrorLog.Println("Error while playing track:", err)
			m.sendPlaybackError(trackState, "A Fatal error occurred while playing the track. Please mention **@Pengasuh Anak** to report this issue.")
			m.queueMu.Lock()
			m.finishLocked()
			m.queueMu.Unlock()
			m.disconnectFromVoiceChannel()
			return
		}

		m.queueMu.Lock()
		m.playCtxCancel = nil
		m.queueMu.Unlock()
	}
}
