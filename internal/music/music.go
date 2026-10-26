package music

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/snowflake/v2"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/model"
	"github.com/taufiq30s/chisa/internal/music/providers"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

// MusicService defines the external interface for the music feature.
// All handlers interact with MusicBot only through this interface.
type MusicService interface {
	Load(s *discordgo.Session, i *discordgo.InteractionCreate, data string)
	Skip(s *discordgo.Session, i *discordgo.InteractionCreate)
	Stop(s *discordgo.Session, i *discordgo.InteractionCreate)
	Disconnect(s *discordgo.Session, i *discordgo.InteractionCreate)
	ClearQueueNow(s *discordgo.Session, i *discordgo.InteractionCreate)
	Pause(s *discordgo.Session, i *discordgo.InteractionCreate)
	Resume(s *discordgo.Session, i *discordgo.InteractionCreate)
	// HandleVoiceLeave is the no-interaction teardown used by the voice state
	// event when Discord pulls the bot out of the channel.
	HandleVoiceLeave()
	ClearQueue()
	SearchNextPage(s *discordgo.Session, i *discordgo.InteractionCreate)
	SearchPreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate)
	SearchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, trackID string)
	// ListQueue renders the current queue with Prev/Next pagination;
	// ListQueueNextPage/PreviousPage are the button handlers for that panel.
	ListQueue(s *discordgo.Session, i *discordgo.InteractionCreate)
	ListQueueNextPage(s *discordgo.Session, i *discordgo.InteractionCreate)
	ListQueuePreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate)
	Shutdown()
}

type MusicBot struct {
	queue         []trackState
	resolver      *Resolver
	streamer      *Streamer
	searchResults map[string]*searchResultState
	searchMu      sync.RWMutex
	// queueLists holds /music list pagination state, keyed by user ID. It is
	// deliberately a separate map+mutex from searchResults: the queue is
	// rewritten by the player goroutine while a search panel may still be open,
	// and mixing them would make one panel's sweeper delete the other's message.
	queueLists  map[string]*queueListState
	queueListMu sync.Mutex
	// queueMu guards queue, isPlaying, playCtxCancel, playerDone and
	// stopRequested. Every read or write of any of those needs it - the player
	// goroutine and the interaction goroutines touch them concurrently.
	queueMu sync.Mutex
	// streamerMu guards the streamer pointer. It is written when joining or
	// leaving a voice channel (interaction goroutines) and read by the player
	// goroutine, so it cannot share queueMu without risking a nil deref.
	streamerMu    sync.RWMutex
	isPlaying     bool
	playCtxCancel context.CancelFunc
	// playerDone is closed by the player goroutine when it exits. stopPlayback
	// waits on it, which is what guarantees no second player is ever spawned
	// while the previous one is still cleaning up.
	playerDone    chan struct{}
	stopRequested bool
	// currentTrack is the track the player goroutine is playing right now. The
	// playing track is otherwise only a local variable inside playerLoop, so
	// /music list could not show it. Nil when nothing is playing.
	currentTrack *trackState
	// lastChannelId remembers the channel the most recent track was queued
	// from. Unlike currentTrack it is NOT cleared when playback ends: the idle
	// auto-disconnect needs somewhere to post its "No Activity" message after
	// the queue has already drained.
	lastChannelId string
	// idleTimer fires the auto-disconnect after AutoDisconnectDelay. idleToken
	// is a generation counter: time.Timer.Stop can report success while the
	// callback is already running, so the callback re-checks the token and does
	// nothing if the timer has been rescheduled or cancelled since.
	idleTimer *time.Timer
	idleToken uint64

	session      *discordgo.Session
	guildID      snowflake.ID
	featureName  string
	shutdown     chan struct{}
	shutdownOnce sync.Once
}

func New(s *discordgo.Session, botId string, guildId string, rdb *redis.Client) (*MusicBot, error) {
	yt, err := providers.NewYoutubeProvider()
	if err != nil {
		return nil, err
	}
	bot := &MusicBot{
		queue:         make([]trackState, InitialQueueCapacity),
		resolver:      &Resolver{providers: []AudioProvider{yt}, cache: newSearchCache(rdb)},
		streamer:      nil,
		searchResults: make(map[string]*searchResultState),
		queueLists:    make(map[string]*queueListState),
		session:       s,
		guildID:       snowflake.MustParse(guildId),
		featureName:   "Chisa Music Player",
		isPlaying:     false,
		shutdown:      make(chan struct{}),
	}
	// Search sessions hold a reference to their Discord message; without this
	// sweep the map grows for the life of the process and expired pickers stay
	// in the channel.
	go bot.InitializeCleanSearchCache()
	return bot, nil
}

func formatDuration(seconds int) string {
	hours := seconds / (MinutesPerHour * SecondsPerMinute)
	minutes := seconds / SecondsPerMinute % MinutesPerHour
	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds%SecondsPerMinute)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds%SecondsPerMinute)
}

func (m *MusicBot) generateMusicInformation(state trackState, status string) {
	// Snapshot shared playback state under the lock, then build the embed from
	// the snapshot. The player goroutine mutates the queue concurrently, so
	// calling len(m.queue) directly here would be an unsynchronized read.
	m.queueMu.Lock()
	queueLength := len(m.queue)
	m.queueMu.Unlock()

	body := responses.CreateMessageEmbed(
		m.session, state.track.Title,
		"",
		fmt.Sprintf("%s (%s)", m.featureName, status),
		responses.SetThumbnailUrl(state.track.Thumbnail),
		responses.SetUrl(state.track.Url),
		responses.SetColor("5e11d9"),
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name:   "Added by",
				Value:  *state.addedBy,
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  formatDuration(state.track.Length),
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  state.track.Provider,
				Inline: true,
			},
			{
				Name:   "Num of Tracks in Queue",
				Value:  fmt.Sprintf("%d", queueLength),
				Inline: true,
			},
		}),
	)
	// Two ways to answer, and picking the wrong one loses the message.
	//
	// SendDefer edits the interaction's deferred reply, which is only valid
	// while that reply is still unanswered. "Added to Queue" is always that
	// case; "Now Playing" is too, but *only* for the track that ownsDefer (the
	// one queued while nothing was playing). Any later "Now Playing" belongs to
	// the player goroutine after the previous track finished - that
	// interaction is long answered or expired, so post to the channel instead.
	if state.interaction != nil && (status != "Now Playing" || state.ownsDefer) {
		err := responses.InteractionResponse(m.session, state.interaction.Interaction).WithEmbed(body).SendDefer()
		if err != nil {
			utils.ErrorLog.Println("Failed to send interaction response", err)
		}
		return
	}

	// No interaction to answer (the player goroutine owns this update), so post
	// to the channel the track was queued from. channelId is nil for tracks
	// restored without an originating interaction.
	if state.channelId == nil {
		return
	}
	if _, err := m.session.ChannelMessageSendEmbed(*state.channelId, body); err != nil {
		utils.ErrorLog.Println("Failed to send message", err)
	}
}

// generatePlaylistInformation answers a playlist import with a single summary
// card. The full track list is deliberately omitted: the queue can hold dozens
// of entries, and /music list already pages through it.
//
// It always uses SendDefer because the summary card is what answers this
// command's deferred reply (see addPlaylistToQueue).
func (m *MusicBot) generatePlaylistInformation(state trackState, tracks []model.Track) {
	m.queueMu.Lock()
	queueLength := len(m.queue)
	m.queueMu.Unlock()

	totalLength := 0
	for _, track := range tracks {
		totalLength += track.Length
	}

	body := responses.CreateMessageEmbed(
		m.session, fmt.Sprintf("Playlist - %d tracks", len(tracks)),
		"",
		fmt.Sprintf("%s (%s)", m.featureName, "Added to Queue"),
		responses.SetThumbnailUrl(state.track.Thumbnail),
		responses.SetUrl(state.track.Url),
		responses.SetColor("5e11d9"),
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name:   "Added by",
				Value:  *state.addedBy,
				Inline: true,
			},
			{
				Name:   "Tracks",
				Value:  fmt.Sprintf("%d", len(tracks)),
				Inline: true,
			},
			{
				Name:   "Total Duration",
				Value:  formatDuration(totalLength),
				Inline: true,
			},
			{
				Name:   "Num of Tracks in Queue",
				Value:  fmt.Sprintf("%d", queueLength),
				Inline: true,
			},
		}),
	)
	if err := responses.InteractionResponse(m.session, state.interaction.Interaction).WithEmbed(body).SendDefer(); err != nil {
		utils.ErrorLog.Println("Failed to send playlist response", err)
	}
}

// Shutdown gracefully stops the music bot and cleans up resources.
// Safe to call more than once: shutdownOnce prevents the "close of closed
// channel" panic you would otherwise get when two shutdown paths race
// (SIGINT handled twice, or Shutdown plus a deferred cleanup).
func (m *MusicBot) Shutdown() {
	m.shutdownOnce.Do(func() {
		m.cancelIdleDisconnect()
		m.stopPlayback()
		close(m.shutdown)
	})

	m.ClearQueue()
	m.disconnectFromVoiceChannel()

	// Clear search results
	m.searchMu.Lock()
	for _, v := range m.searchResults {
		if v.channelId != "" && v.messageId != "" {
			_ = m.session.ChannelMessageDelete(v.channelId, v.messageId)
		}
	}
	m.searchResults = make(map[string]*searchResultState)
	m.searchMu.Unlock()
}
