package music

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/snowflake/v2"
	"github.com/taufiq30s/chisa/internal/responses"
)

func (MusicBot *MusicBot) getPlatform(url string) string {
	for platform, prefixs := range MusicBot.supportedPlatformsPrefix {
		for _, prefix := range prefixs {
			if strings.HasPrefix(url, prefix) {
				return platform
			}
		}
	}
	return ""
}

func (MusicBot *MusicBot) Play(s *discordgo.Session, i *discordgo.InteractionCreate, data string) {
	_, err := url.ParseRequestURI(data)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     MusicBot.featureName,
			Title:       "Invalid Link or Unsupport Media Provider",
			Description: "This feature only support Youtube and Spotify. Please check again!",
		}).Execute()
		return
	}
	switch platform := MusicBot.getPlatform(data); platform {
	case "spotify":
		MusicBot.playSpotify(s, i, data)
	case "youtube":
		MusicBot.playYoutubeTest(s, i, data)
	default:
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     MusicBot.featureName,
			Title:       "Unsupport Media Provider",
			Description: "This feature only support Spotify. Please check again!",
		}).Execute()
	}
}

func (MusicBot *MusicBot) playYoutubeTest(s *discordgo.Session, i *discordgo.InteractionCreate, url string) {
	identifier := lavalink.SearchTypeYouTube.Apply(url)
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error while getting voice state: `%s`", err),
			},
		})
	}

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error while getting voice state: `%s`", err),
			},
		})
	}

	player := MusicBot.Client.Lavalink.Player(snowflake.MustParse(i.GuildID))
	// queue := MusicBot.Queue.Get(event.GuildID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var toPlay *lavalink.Track
	MusicBot.Client.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(track lavalink.Track) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loading track: [`%s`](<%s>)", track.Info.Title, *track.Info.URI)),
			})
			if player.Track() == nil {
				toPlay = &track
			} else {
				fmt.Println(track)
			}
		},
		func(playlist lavalink.Playlist) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loaded playlist: `%s` with `%d` tracks", playlist.Info.Name, len(playlist.Tracks))),
			})
		},
		func(tracks []lavalink.Track) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Loaded search result: [`%s`](<%s>)", tracks[0].Info.Title, *tracks[0].Info.URI)),
			})
			if player.Track() == nil {
				toPlay = &tracks[0]
			} else {
				fmt.Println(tracks[0])
			}
		},
		func() {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Nothing found for: `%s`", identifier)),
			})
		},
		func(err error) {
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: json.Ptr(fmt.Sprintf("Error while looking up query: `%s`", err)),
			})
		}))

	if err := s.ChannelVoiceJoinManual(i.GuildID, voiceState.ChannelID, false, false); err != nil {
		print(err)
	}

	fmt.Println("ok")
	fmt.Println(player.Node().Status())
	err = player.Update(context.TODO(), lavalink.WithTrack(*toPlay))
	if err != nil {
		fmt.Println(err)
	}
}

func (MusicBot *MusicBot) playSpotify(s *discordgo.Session, i *discordgo.InteractionCreate, url string) {
	// Get Track ID
	trackId := url[strings.LastIndex(url, "/")+1 : strings.Index(url, "?")]
	metadata, err := MusicBot.Client.SpotifyClient.GetTrack(&trackId)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature: MusicBot.featureName,
			Title:   "Failed to get Spotify Metadata",
			Err:     err,
		}).Execute()
		return
	}

	messageEmbed := responses.CreateMessageEmbed(
		s,
		metadata.Name,
		"",
		MusicBot.featureName+" (Added to Queue)",
		responses.SetThumbnailUrl(metadata.Album.Images[0].Url),
		responses.SetUrl(url),
		responses.SetColor("5e11d9"),
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name:   "Added by",
				Value:  i.Member.User.Username,
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  strconv.Itoa(metadata.DurationMs/(1000*60)) + ":" + strconv.Itoa(metadata.DurationMs/1000%60),
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  "Spotify",
				Inline: true,
			},
			{
				Name:   "Queue Length",
				Value:  fmt.Sprintf("%d", 0),
				Inline: true,
			},
		}),
	)
	responses.InteractionResponse(
		s, i.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		messageEmbed,
	).Execute()
}

// TODO: Because there are problem with responses validation. This feature will be postpone
// func playYoutube(state *responsesState, url string) {
// 	ytClient := youtube.Client{}

// 	// Get Metadata
// 	meta, err := ytClient.GetVideo(url)
// 	if err != nil {
// 		responses.ErrorResponse(state.chisa.Session, state.interaction, &responses.ErrorResponseData{
// 			Feature: featureName,
// 			Title:   "Failed to fetch metadata",
// 			Err:     err,
// 		}).Execute()
// 		return
// 	}

// 	if meta != nil {
// 		state.chisa.Session.InteractionRespond(
// 			state.interaction.Interaction,
// 			&discordgo.InteractionResponse{
// 				Type: discordgo.InteractionResponseChannelMessageWithSource,
// 				Data: &discordgo.InteractionResponseData{
// 					Embeds: []*discordgo.MessageEmbed{
// 						responses.CreateMessageEmbed(
// 							state.chisa.Session,
// 							meta.Title,
// 							"",
// 							featureName+" (Added to Queue)",
// 							responses.SetThumbnailUrl(meta.Thumbnails[0].URL),
// 							responses.SetUrl(url),
// 							responses.SetColor("5e11d9"),
// 							responses.SetFields([]*discordgo.MessageEmbedField{
// 								{
// 									Name:   "Added by",
// 									Value:  state.interaction.Member.User.Username,
// 									Inline: true,
// 								},
// 								{
// 									Name:   "Duration",
// 									Value:  meta.Duration.String(),
// 									Inline: true,
// 								},
// 								{
// 									Name:   "Provider",
// 									Value:  "Youtube",
// 									Inline: true,
// 								},
// 								{
// 									Name:   "Queue Length",
// 									Value:  fmt.Sprintf("%d", 0),
// 									Inline: true,
// 								},
// 							}),
// 						),
// 					},
// 				},
// 			},
// 		)
// 	}
// }
