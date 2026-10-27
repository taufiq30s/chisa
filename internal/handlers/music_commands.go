package handlers

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

var (
	musicCommandHandler = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		// Acknowledge once, up front, for every music subcommand. The handlers
		// below answer with SendDefer/ExecuteDefer, which EDIT this deferred
		// response. Never call Defer() again further down: Discord allows only
		// one acknowledgement per interaction and a second one fails with
		// "40060 Interaction has already been acknowledged", after which the
		// user sees nothing at all.
		if err := responses.InteractionResponse(chisa.Session, interaction.Interaction).Defer(); err != nil {
			utils.ErrorLog.Println(err)
			return
		}

		switch options := interaction.ApplicationCommandData().Options; options[OptionIndexFirst].Name {
		case "play":
			chisa.Music.Load(
				chisa.Session,
				interaction,
				parseIdentifier(options[OptionIndexFirst].Options[OptionIndexFirst].StringValue()),
			)
		case "skip":
			chisa.Music.Skip(chisa.Session, interaction)
		case "stop":
			chisa.Music.Stop(chisa.Session, interaction)
		case "disconnect":
			chisa.Music.Disconnect(chisa.Session, interaction)
		case "clear":
			chisa.Music.ClearQueueNow(chisa.Session, interaction)
		case "pause":
			chisa.Music.Pause(chisa.Session, interaction)
		case "resume":
			chisa.Music.Resume(chisa.Session, interaction)
		case "list":
			chisa.Music.ListQueue(chisa.Session, interaction)
		}
	}
	musicCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "music",
			Description: "Music Player",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "play",
					Description: "Play a music",
					Type:        1,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "query",
							Description: "Name of song or Music Platform URL",
							Required:    true,
						},
					},
				},
				{
					Name:        "pause",
					Description: "Pause current song",
					Type:        1,
				},
				{
					Name:        "resume",
					Description: "Resume current song",
					Type:        1,
				},
				{
					Name:        "skip",
					Description: "Skip current song",
					Type:        1,
				},
				{
					Name:        "stop",
					Description: "Stop music player",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "disconnect",
					Description: "Disconnect from voice channel",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "clear",
					Description: "Remove every queued track (current track keeps playing)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "list",
					Description: "Show the current queue with pagination",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}
)

func parseIdentifier(data string) string {
	parsedURL, err := url.ParseRequestURI(data)
	if err != nil {
		// Not a URL at all, so it is a search phrase. Return it untouched:
		// the youtube provider decides how to search (yt-dlp's ytsearch).
		// Prefixing it here would make providers.IsURL treat it as a URL and
		// every search would fail with "no provider found".
		return data
	}

	// Handle YouTube URLs
	if strings.Contains(parsedURL.Host, "youtube.com") || strings.Contains(parsedURL.Host, "youtu.be") {
		if parsedURL.Host == "youtu.be" {
			return fmt.Sprintf("https://youtu.be/%s", strings.Trim(parsedURL.Path, "/"))
		}
		if parsedURL.Path == "/watch" && parsedURL.Query().Has("v") {
			return fmt.Sprintf("https://www.youtube.com/watch?v=%s", parsedURL.Query().Get("v"))
		}
		videoID := strings.Trim(parsedURL.Path, "/")
		if strings.HasPrefix(parsedURL.Path, "/embed/") {
			return fmt.Sprintf("https://www.youtube.com/embed/%s", strings.TrimPrefix(videoID, "embed/"))
		}
		if strings.HasPrefix(parsedURL.Path, "/v/") {
			return fmt.Sprintf("https://www.youtube.com/v/%s", strings.TrimPrefix(videoID, "v/"))
		}
	}

	// Handle Spotify track URLs
	if parsedURL.Host == "open.spotify.com" && strings.HasPrefix(parsedURL.Path, "/track/") {
		trackID := strings.TrimPrefix(parsedURL.Path, "/track/")
		return fmt.Sprintf("https://open.spotify.com/track/%s", trackID)
	}
	return data
}
