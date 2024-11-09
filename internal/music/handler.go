package music

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func HandleSearchNextPage(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{}) {
	params[0].(*MusicBot).moveSearchPage(s, i, true)
}

func HandleSearchPreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{}) {
	params[0].(*MusicBot).moveSearchPage(s, i, false)
}

func HandleSearchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, params ...interface{}) {
	data := i.MessageComponentData()
	params[0].(*MusicBot).selectSearchResult(s, i, data.Values[0])
}

func (m *MusicBot) onTrackStart(player disgolink.Player, event lavalink.TrackStartEvent) {
	fmt.Println("Player start")
	trackState := m.getFirstTrack()
	m.generateMusicInformation(
		trackState,
		"Now Playing",
	)
}

func (m *MusicBot) onTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	m.removeFromQueue()
	fmt.Println("Player end")
	trackState := m.getFirstTrack()
	if trackState == nil {
		return
	}
	player.Update(context.TODO(), lavalink.WithTrack(*trackState.lavaTrack))
}

func (m *MusicBot) onTrackStuck(player disgolink.Player, event lavalink.TrackStuckEvent) {
	fmt.Println("Track Stuck")
}

func (m *MusicBot) onWebSocketClosed(player disgolink.Player, event lavalink.WebSocketClosedEvent) {
	fmt.Println("WebSocket Closed")
}
