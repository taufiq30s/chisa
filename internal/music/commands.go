package music

import (
	"context"
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
	err := responses.InteractionResponse(s, i.Interaction).Defer()
	if err != nil {
		utils.ErrorLog.Println(err)
		return
	}
	if m == nil {
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

	m.Client.BestNode().LoadTracksHandler(ctx, data, disgolink.NewResultHandler(
		func(track lavalink.Track) {
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
			}).ExecuteDefer()
		},
		func(err error) {
			responses.ErrorResponse(s, i, &responses.ErrorResponseData{
				Feature:     m.featureName,
				Title:       "Invalid URL",
				Description: "The provider URL is invalid.",
			}).ExecuteDefer()
			utils.ErrorLog.Println(err.Error())
		}))

	if playingTrack == nil {
		return
	}

	if err := s.ChannelVoiceJoinManual(i.GuildID, voiceState.ChannelID, false, false); err != nil {
		utils.ErrorLog.Printf("failed to join voice channel: %v\n", err)
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "Failed to join voice channel",
			Description: err.Error(),
		}).ExecuteDefer()
		return
	}
	m.Play(playingTrack)
}

func (m *MusicBot) Play(track *lavalink.Track) {
	utils.InfoLog.Println("Player Node: ", m.player.Node().Config().Name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.player.Update(ctx, lavalink.WithTrack(*track))
	if err != nil {
		utils.ErrorLog.Println("failed to update player:", err)
	}
}

func (m *MusicBot) Skip(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() == nil {
		err := responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     m.featureName,
			Title:       "No track playing",
			Description: "There is no track playing.",
		}).Execute()
		if err != nil {
			utils.ErrorLog.Println("Error when create error response", err)
		}
		return
	}

	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(
		responses.CreateMessageEmbed(
			s,
			"Skipped",
			"Track Skipped.",
			m.featureName,
			responses.SetColor("0bdd47"),
		)).Send()
	m.setNextTrackAsInteractionReply(i)
	player.Update(context.Background(), lavalink.WithPosition(player.Track().Info.Length))
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
	player.Update(context.Background(), lavalink.WithPaused(true))
	embed := responses.CreateMessageEmbed(
		s, "Paused",
		"Paused.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(embed).Send()
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
	player.Update(context.Background(), lavalink.WithPaused(false))
	embed := responses.CreateMessageEmbed(
		s, "Resumed",
		"Resumed.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(embed).Send()
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
	player.Update(context.Background(), lavalink.WithNullTrack())
	embed := responses.CreateMessageEmbed(
		s, "Stopped playing",
		"Stopped playing.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(embed).Send()
}

func (m *MusicBot) Disconnect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	player := m.Client.Player(snowflake.MustParse(i.GuildID))
	if player.Track() != nil {
		player.Update(context.Background(), lavalink.WithNullTrack())
	}
	s.ChannelVoiceJoinManual(i.GuildID, "", false, false)
	embed := responses.CreateMessageEmbed(
		s, "Disconnected",
		"Disconnected.",
		m.featureName,
		responses.SetColor("0bdd47"),
	)
	responses.InteractionResponse(
		s, i.Interaction,
	).WithEmbed(embed).Send()
}
