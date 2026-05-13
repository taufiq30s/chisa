package music

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

func (m *MusicBot) SearchNextPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveSearchPage(s, i, true)
}

func (m *MusicBot) SearchPreviousPage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	m.moveSearchPage(s, i, false)
}

func (m *MusicBot) SearchSelect(s *discordgo.Session, i *discordgo.InteractionCreate, trackID string) {
	m.selectSearchResult(s, i, trackID)
}

// ForwardVoiceServerUpdate forwards a Discord voice server update to the Lavalink client.
func (m *MusicBot) ForwardVoiceServerUpdate(ctx context.Context, guildID snowflake.ID, token string, endpoint string) {
	m.Client.OnVoiceServerUpdate(ctx, guildID, token, endpoint)
}

// ForwardVoiceStateUpdate forwards a Discord voice state update to the Lavalink client.
func (m *MusicBot) ForwardVoiceStateUpdate(ctx context.Context, guildID snowflake.ID, channelID *snowflake.ID, sessionID string) {
	m.Client.OnVoiceStateUpdate(ctx, guildID, channelID, sessionID)
}

func (m *MusicBot) onPlayerUpdate(player disgolink.Player, _ lavalink.PlayerUpdateMessage) {
	position := player.Position()
	positionInSeconds := position.Seconds()
	if m.lastPosition == positionInSeconds && !player.Paused() {
		track := player.Track()
		if track != nil {
			guildId, err := utils.GetEnv("AKASHIC_SERVER_ID")
			if err != nil {
				utils.ErrorLog.Fatalln(err)
				return
			}
			player.Update(context.TODO(), lavalink.WithNullTrack())
			player = m.Client.Player(snowflake.MustParse(guildId))
			player.Update(context.TODO(), lavalink.WithTrack(*track), lavalink.WithPosition(position))
		}
	}
}

func (m *MusicBot) onTrackStart(player disgolink.Player, event lavalink.TrackStartEvent) {
	trackState := m.getFirstTrack()
	m.generateMusicInformation(
		trackState,
		"Now Playing",
	)
}

func (m *MusicBot) onTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	channelId := m.getFirstTrack().channelId
	m.removeFromQueue()
	fmt.Println("Player end")
	trackState := m.getFirstTrack()
	if trackState == nil {
		_, err := m.session.ChannelMessageSendEmbed(*channelId, responses.CreateMessageEmbed(
			m.session,
			"No Song",
			"No song in queue",
			m.featureName,
			responses.SetColor("0bdd47"),
		))
		if err != nil {
			utils.ErrorLog.Printf("Error sending message: %v", err)
		}
		return
	}
	guildId, err := utils.GetEnv("AKASHIC_SERVER_ID")
	if err != nil {
		utils.ErrorLog.Fatalln(err)
		return
	}
	m.player = m.Client.Player(snowflake.MustParse(guildId))
	m.Play(trackState.lavaTrack)
}

func (m *MusicBot) onTrackException(player disgolink.Player, event lavalink.TrackExceptionEvent) {
	fmt.Println(event.Exception.Error())
	utils.ErrorLog.Println(event.Exception.Error())
}

func (m *MusicBot) onTrackStuck(player disgolink.Player, event lavalink.TrackStuckEvent) {
	fmt.Println("Track Stuck")
}

func (m *MusicBot) onWebSocketClosed(player disgolink.Player, event lavalink.WebSocketClosedEvent) {
	fmt.Println("WebSocket Closed")
}
