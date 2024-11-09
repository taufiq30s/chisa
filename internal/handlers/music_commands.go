package handlers

import (
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/music"
)

var (
	musicCommandHandler = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[0].Name {
		case "play":
			chisa.Music.Play(
				chisa.Session,
				interaction,
				parseIdentifier(options[0].Options[0].StringValue()),
			)
		case "skip":
			chisa.Music.Skip(chisa.Session, interaction)
		case "stop":
			// stop(chisa.Session, interaction)
		case "disconnect":
			// disconnect(chisa.Session, interaction)
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
					Name:        "skip",
					Description: "Skip current song",
					Type:        1,
				},
				{
					Name:        "stop",
					Description: "Stop music player",
					Type:        1,
				},
				{
					Name:        "disconnect",
					Description: "Disconnect from voice channel",
					Type:        1,
				},
			},
		},
	}
	SearchButtonHandler = map[string]componentFunction{
		"search-next":     music.HandleSearchNextPage,
		"search-previous": music.HandleSearchPreviousPage,
	}
	SearchSelectHandler = map[string]componentFunction{
		"search-select": music.HandleSearchSelect,
	}
)

func parseIdentifier(data string) string {
	_, err := url.ParseRequestURI(data)
	if err != nil {
		return lavalink.SearchTypeYouTube.Apply(data)
	}
	return data
}
