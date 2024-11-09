package music

import (
	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

type trackState struct {
	lavaTrack   *lavalink.Track
	interaction *discordgo.InteractionCreate
	addedBy     string
	channelId   string
}

func (m *MusicBot) addToQueue(t *lavalink.Track, i *discordgo.InteractionCreate) {
	state := trackState{
		lavaTrack: t,
		addedBy:   i.Member.User.Username,
		channelId: i.ChannelID,
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
			addedBy:   i.Member.User.Username,
			channelId: i.ChannelID,
		}
		m.queue = append(m.queue, state)
	}

	m.generateMusicInformation(
		&trackState{
			lavaTrack:   &p.Tracks[0],
			interaction: i,
			addedBy:     i.Member.User.Username,
			channelId:   i.ChannelID,
		},
		"Added Playlist to Queue",
	)
}

func (m *MusicBot) getFirstTrack() *trackState {
	if len(m.queue) == 0 {
		return nil
	}
	return &m.queue[0]
}

func (m *MusicBot) setNextTrackAsInteractionReply(i *discordgo.InteractionCreate) {
	if len(m.queue) == 0 {
		return
	}
	if len(m.queue) == 1 {
		m.queue[0].interaction = i
		return
	}
	m.queue[1].interaction = i
}

func (m *MusicBot) removeFromQueue() {
	if len(m.queue) == 0 {
		return
	}
	m.queue = m.queue[1:]
}

func (m *MusicBot) clearQueue() {
	m.queue = []trackState{}
}
