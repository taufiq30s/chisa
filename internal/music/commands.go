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
func (m *MusicBot) Load(s *discordgo.Session, i *discordgo.InteractionCreate, data string) {
	if m == nil {
		fmt.Println("MusicBot Client is not ready.")
		utils.ErrorLog.Println("MusicBot Client is not ready.")
		return
	}
	if m.Client.BestNode() == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No available nodes",
			Description: "No available nodes to play music. Please contact administrator.",
		}).Execute()
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

	// Find track and store selected track
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var playingTrack *lavalink.Track
	defer cancel()

	fmt.Println("Player Node: ", m.Client.BestNode().Config().Name)
	m.Client.BestNode().LoadTracksHandler(ctx, data, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			fmt.Println("Start 1")
			if m.player.Track() == nil {
				playingTrack = &track
			}
			m.addToQueue(&track, i)
		},
		func(playlist lavalink.Playlist) {
			if m.player.Track() == nil {
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
		utils.ErrorLog.Println(err)
	}
	m.Play(playingTrack)
}

func (m *MusicBot) Play(track *lavalink.Track) {
	utils.InfoLog.Println("Player Node: ", m.player.Node().Config().Name)
	fmt.Println("Player Node: ", m.player.Node().Config().Name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.player.Update(ctx, lavalink.WithTrack(*track))
	if err != nil {
		fmt.Println(err)
		utils.ErrorLog.Println(err)
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

func (m *MusicBot) Pause(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No track playing",
			Description: "There is no track playing.",
		}).Execute()
		return
	}
	player.Update(context.TODO(), lavalink.WithPaused(true))
	resp := responses.CreateMessageEmbed(
		s, "Paused",
		"Paused.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		resp,
	).Execute()
}

func (m *MusicBot) Resume(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No track playing",
			Description: "There is no track playing.",
		}).Execute()
		return
	}
	player.Update(context.TODO(), lavalink.WithPaused(false))
	resp := responses.CreateMessageEmbed(
		s, "Resumed",
		"Resumed.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		resp,
	).Execute()
}

func (m *MusicBot) Stop(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() == nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No track playing",
			Description: "There is no track playing.",
		}).Execute()
		return
	}
	m.clearQueue()
	player.Update(context.TODO(), lavalink.WithNullTrack())
	resp := responses.CreateMessageEmbed(
		s, "Stopped playing",
		"Stopped playing.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		resp,
	).Execute()
}

func (m *MusicBot) Disconnect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() != nil {
		player.Update(context.TODO(), lavalink.WithNullTrack())
	}
	s.ChannelVoiceJoinManual(i.GuildID, "", false, false)
	resp := responses.CreateMessageEmbed(
		s, "Disconnected",
		"Disconnected.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		resp,
	).Execute()
}
