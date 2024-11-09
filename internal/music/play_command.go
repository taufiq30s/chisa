package music

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

/*
Play Command
*/
func (m *MusicBot) Play(s *discordgo.Session, i *discordgo.InteractionCreate, data string) {
	if m == nil {
		fmt.Println("MusicBot Client is not ready.")
		utils.ErrorLog.Println("MusicBot Client is not ready.")
		return
	}
	// Create voicestate
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Error while getting voice state",
			Description: err.Error(),
		}).Execute()
	}

	// Initialize Player
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find track and store selected track
	var playingTrack *lavalink.Track
	m.Client.BestNode().LoadTracksHandler(ctx, data, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			fmt.Println("Start 1")
			if player.Track() == nil {
				playingTrack = &track
			}
			m.addToQueue(&track, i)
		},
		func(playlist lavalink.Playlist) {
			if player.Track() == nil {
				playingTrack = &playlist.Tracks[0]
			}
			m.addPlaylistToQueue(&playlist, i)
		},
		func(tracks []lavalink.Track) {
			m.search(s, i, tracks)
		},
		func() {
			responses.ErrorResponse(s, i, &responses.ErrorResponseData{
				Feature:     m.featureName,
				Title:       "No tracks found",
				Description: "No tracks found for: " + data,
			}).Execute()
		},
		func(err error) {
			responses.ErrorResponse(s, i, &responses.ErrorResponseData{
				Feature:     m.featureName,
				Title:       "Invalid URL",
				Description: "The provider URL is invalid.",
			}).Execute()
			utils.ErrorLog.Println(err.Error())
			fmt.Println(err)
		}))

	if playingTrack == nil {
		return
	}

	if err := s.ChannelVoiceJoinManual(i.GuildID, voiceState.ChannelID, false, false); err != nil {
		fmt.Println(err)
	}
	fmt.Println(player.Node().Config().Name)
	err = player.Update(context.TODO(), lavalink.WithTrack(*playingTrack))
	if err != nil {
		fmt.Println(err)
	}
}

func (m *MusicBot) Skip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No track playing",
			Description: "There is no track playing.",
		}).Execute()
		return
	}
	m.setNextTrackAsInteractionReply(i)
	player.Update(context.TODO(), lavalink.WithPosition(player.Track().Info.Length))
}
