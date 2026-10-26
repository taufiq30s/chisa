package music

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/model"
)

// newTestBot is a MusicBot with only the fields the lock-and-queue helpers
// touch: no Discord session, no voice connection, no Redis. Anything that would
// reach the network must not be called from these tests.
func newTestBot() *MusicBot {
	return &MusicBot{queue: []trackState{}, shutdown: make(chan struct{})}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		seconds int
		want    string
	}{
		{0, "00:00"},
		{5, "00:05"},
		{59, "00:59"},
		{60, "01:00"},
		{131, "02:11"}, // the value that used to render as 00:00
		{599, "09:59"},
		{3725, "01:02:05"},
		{36000, "10:00:00"},
	}
	for _, c := range cases {
		if got := formatDuration(c.seconds); got != c.want {
			t.Errorf("formatDuration(%d) = %q, want %q", c.seconds, got, c.want)
		}
	}
}

// The queue is appended by interaction goroutines and drained by the player
// goroutine at the same time. Run with -race: a missing lock shows up as a DATA
// RACE report here instead of a garbled queue in production.
func TestQueueConcurrentAccess(t *testing.T) {
	m := newTestBot()
	state := trackState{}

	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 200; n++ {
				m.queueMu.Lock()
				m.queue = append(m.queue, state)
				m.queueMu.Unlock()
			}
		}()
	}
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 400; n++ {
				m.queueMu.Lock()
				m.dequeueLocked()
				m.queueMu.Unlock()
			}
		}()
	}
	wg.Wait()

	m.ClearQueue()
	if len(m.queue) != 0 {
		t.Fatalf("queue not empty after ClearQueue: %d", len(m.queue))
	}
}

// stopPlayback must not return before the player goroutine has exited. If this
// regresses, a /music play fired right after /music stop spawns a second player
// loop and two goroutines write to the same OpusSend channel.
func TestStopPlaybackWaitsForPlayer(t *testing.T) {
	m := newTestBot()

	done := make(chan struct{})
	m.playerDone = done

	loopDone := make(chan struct{})
	go func() {
		defer close(loopDone)
		time.Sleep(50 * time.Millisecond) // pretend Streamer.Play is unwinding
		close(done)
	}()

	m.stopPlayback()

	select {
	case <-loopDone:
	default:
		t.Fatal("stopPlayback returned before the player goroutine finished")
	}
	if !m.stopRequested {
		t.Fatal("stopPlayback did not set stopRequested")
	}
}

// A stop request must survive until the loop observes it, even when the loop
// was not running at the time (nothing playing). Otherwise the flag leaks into
// the next playback session and kills it immediately.
//
// Tests beginPlaybackLocked rather than startPlayerLoopIfNeeded: the latter
// spawns a goroutine that drains the empty queue and rewrites these same fields
// the instant the lock is released, which a -race run reports as a data race in
// the test itself.
func TestStartPlayerLoopClearsStopRequest(t *testing.T) {
	m := newTestBot()

	m.queueMu.Lock()
	m.stopRequested = true
	_, ok := m.beginPlaybackLocked()
	staleStop := m.stopRequested
	playing := m.isPlaying
	m.queueMu.Unlock()

	if !ok {
		t.Fatal("beginPlaybackLocked refused to start when nothing was playing")
	}
	if staleStop {
		t.Fatal("starting the player kept a stale stopRequested")
	}
	if !playing {
		t.Fatal("starting the player did not mark isPlaying")
	}
}

// The voice state event fires HandleVoiceLeave on every "you left" update, and
// Discord may send it twice (kick + explicit disconnect). It must be a no-op
// when nothing is playing, never block forever, and must leave the queue empty.
func TestHandleVoiceLeaveIsIdempotent(t *testing.T) {
	m := newTestBot()
	m.queue = []trackState{{}}

	m.HandleVoiceLeave()
	m.HandleVoiceLeave()

	if len(m.queue) != 0 {
		t.Fatalf("queue not cleared: %d", len(m.queue))
	}
	if m.isPlaying {
		t.Fatal("isPlaying still true after HandleVoiceLeave")
	}
}

// Only network-looking errors are worth retrying; a bad channel or missing
// permission would just fail twice and double the latency. The transient cases
// are the exact strings from the voice-timeout bug reports.
func TestIsTransientVoiceError(t *testing.T) {
	transient := []string{
		"timeout waiting for voice",
		"dial tcp [2606:4700:7::a29f:8aea]:443: connect: network is unreachable",
		"dial tcp: lookup c-sin18.discord.media on 127.0.0.53:53: server misbehaving",
	}
	for _, msg := range transient {
		if !isTransientVoiceError(errors.New(msg)) {
			t.Errorf("expected retryable: %s", msg)
		}
	}

	permanent := []string{
		"no voice channel to join",
		"Missing Access",
		"Unknown Channel",
	}
	for _, msg := range permanent {
		if isTransientVoiceError(errors.New(msg)) {
			t.Errorf("expected not retryable: %s", msg)
		}
	}

	if isTransientVoiceError(nil) {
		t.Error("nil must not be retryable")
	}
}

// A track may only reuse its interaction's deferred reply for "Now Playing" if
// that reply is still unanswered. Slash commands defer up front; the search
// picker answers with a DeferredMessageUpdate and then deletes the message, so
// editing it again would fail and the user would see nothing.
func TestOwesFirstReply(t *testing.T) {
	command := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{Type: discordgo.InteractionApplicationCommand},
	}
	component := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{Type: discordgo.InteractionMessageComponent},
	}

	if !owesFirstReply(command) {
		t.Error("slash command defer must be reusable")
	}
	if owesFirstReply(component) {
		t.Error("component interaction must not be reusable")
	}
	if owesFirstReply(nil) {
		t.Error("nil interaction must not be reusable")
	}
}

// A playlist import must not hand its deferred reply to any track: the summary
// card spends it. If a track also owned the defer, its "Now Playing" would edit
// the same interaction and the summary would be overwritten or lost.
func TestPlaylistStatesNeverOwnDefer(t *testing.T) {
	i := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type:      discordgo.InteractionApplicationCommand,
			ChannelID: "chan",
			Member:    &discordgo.Member{User: &discordgo.User{Username: "tester"}},
		},
	}
	tracks := []model.Track{{ID: "a", Title: "A"}, {ID: "b", Title: "B"}}

	states := playlistStates(i, tracks)
	if len(states) != 2 {
		t.Fatalf("got %d states, want 2", len(states))
	}
	for n, state := range states {
		if state.ownsDefer {
			t.Errorf("state %d ownsDefer = true, want false", n)
		}
		if *state.addedBy != "tester" || *state.channelId != "chan" {
			t.Errorf("state %d attribution = %q/%q", n, *state.addedBy, *state.channelId)
		}
		if state.track.ID != tracks[n].ID {
			t.Errorf("state %d track = %q, want %q (order must be preserved)", n, state.track.ID, tracks[n].ID)
		}
	}
}
