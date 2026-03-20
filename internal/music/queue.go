package music

import (
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

type trackState struct {
	lavaTrack    *lavalink.Track
	interaction  *discordgo.InteractionCreate
	addedBy      *string
	playlistName *string
	channelId    *string
}

func (m *MusicBot) addToQueue(t *lavalink.Track, i *discordgo.InteractionCreate) {
	state := trackState{
		lavaTrack: t,
		addedBy:   &i.Member.User.Username,
		channelId: &i.ChannelID,
	}
	if m.player.Track() == nil {
		state.interaction = i
	}
	m.queue = append(m.queue, state)
	if m.player.Track() != nil {
		state.interaction = i
		m.generateMusicInformation(
			&state,
			"Added to Queue",
		)
	}
}

func (m *MusicBot) addPlaylistToQueue(p *lavalink.Playlist, i *discordgo.InteractionCreate) {
	for _, t := range p.Tracks {
		state := trackState{
			lavaTrack: &t,
			addedBy:   &i.Member.User.Username,
			channelId: &i.ChannelID,
		}
		m.queue = append(m.queue, state)
	}

	m.generateMusicInformation(
		&trackState{
			lavaTrack:    &p.Tracks[0],
			playlistName: &p.Info.Name,
			interaction:  i,
			addedBy:      &i.Member.User.Username,
			channelId:    &i.ChannelID,
		},
		"Added Playlist to Queue",
	)
}

func (m *MusicBot) getFirstTrack() *trackState {
	if len(m.queue) == InitialQueueCapacity {
		return nil
	}
	return &m.queue[FirstQueueIndex]
}

func (m *MusicBot) setNextTrackAsInteractionReply(i *discordgo.InteractionCreate) {
	if len(m.queue) == InitialQueueCapacity {
		return
	}
	if len(m.queue) == QueueOffsetForLength {
		m.queue[FirstQueueIndex].interaction = i
		return
	}
	m.queue[SecondQueueIndex].interaction = i
}

func (m *MusicBot) removeFromQueue() {
	if len(m.queue) == InitialQueueCapacity {
		return
	}
	m.queue = m.queue[QueueOffsetForLength:]
}

func (m *MusicBot) clearQueue() {
	m.queue = []trackState{}
}
