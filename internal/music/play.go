package music

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"
	"github.com/taufiq30s/chisa/internal/bot"
)

func getPlatform(url string) string {
	for platform, prefixs := range supportedPlatformsPrefix {
		for _, prefix := range prefixs {
			if strings.HasPrefix(url, prefix) {
				return platform
			}
		}
	}
	return ""
}

func play(state *botState, data string) {
	_, err := url.ParseRequestURI(data)
	if err != nil {
		bot.ErrorResponse(state.chisa.Session, state.interaction, &bot.ErrorResponseData{
			Feature:     featureName,
			Title:       "Invalid Link or Unsupport Media Provider",
			Description: "This feature only support Youtube and Spotify. Please check again!",
		}).Execute()
		return
	}
	switch platform := getPlatform(data); platform {
	case "spotify":
		playSpotify(state, data)
	default:
		bot.ErrorResponse(state.chisa.Session, state.interaction, &bot.ErrorResponseData{
			Feature:     featureName,
			Title:       "Unsupport Media Provider",
			Description: "This feature only support Spotify. Please check again!",
		}).Execute()
	}
}

func playSpotify(state *botState, url string) {
	// Get Track ID
	trackId := url[strings.LastIndex(url, "/")+1 : strings.Index(url, "?")]
	metadata, err := state.chisa.SpotifySession.GetTrack(&trackId)
	if err != nil {
		bot.ErrorResponse(state.chisa.Session, state.interaction, &bot.ErrorResponseData{
			Feature: featureName,
			Title:   "Failed to get Spotify Metadata",
			Err:     err,
		}).Execute()
		return
	}
	messageEmbed := bot.CreateMessageEmbed(
		state.chisa.Session,
		metadata.Name,
		"",
		featureName+" (Added to Queue)",
		bot.SetThumbnailUrl(metadata.Album.Images[0].Url),
		bot.SetUrl(url),
		bot.SetColor("5e11d9"),
		bot.SetFields([]*discordgo.MessageEmbedField{
			{
				Name:   "Added by",
				Value:  state.interaction.Member.User.Username,
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
	bot.InteractionResponse(
		state.chisa.Session,
		state.interaction.Interaction,
		discordgo.InteractionResponseChannelMessageWithSource,
		false,
		messageEmbed,
	).Execute()
}

// TODO: Because there are problem with bot validation. This feature will be postpone
func playYoutube(state *botState, url string) {
	ytClient := youtube.Client{}

	// Get Metadata
	meta, err := ytClient.GetVideo(url)
	if err != nil {
		bot.ErrorResponse(state.chisa.Session, state.interaction, &bot.ErrorResponseData{
			Feature: featureName,
			Title:   "Failed to fetch metadata",
			Err:     err,
		}).Execute()
		return
	}

	if meta != nil {
		state.chisa.Session.InteractionRespond(
			state.interaction.Interaction,
			&discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
						bot.CreateMessageEmbed(
							state.chisa.Session,
							meta.Title,
							"",
							featureName+" (Added to Queue)",
							bot.SetThumbnailUrl(meta.Thumbnails[0].URL),
							bot.SetUrl(url),
							bot.SetColor("5e11d9"),
							bot.SetFields([]*discordgo.MessageEmbedField{
								{
									Name:   "Added by",
									Value:  state.interaction.Member.User.Username,
									Inline: true,
								},
								{
									Name:   "Duration",
									Value:  meta.Duration.String(),
									Inline: true,
								},
								{
									Name:   "Provider",
									Value:  "Youtube",
									Inline: true,
								},
								{
									Name:   "Queue Length",
									Value:  fmt.Sprintf("%d", 0),
									Inline: true,
								},
							}),
						),
					},
				},
			},
		)
	}
}
