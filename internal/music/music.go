package music

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
)

var (
	featureName              string              = "Chisa Music Player"
	supportedPlatformsPrefix map[string][]string = map[string][]string{
		"youtube": {
			"https://youtube.com/",
			"https://www.youtube.com/",
			"https://www.youtube.com/shorts/",
			"https://youtu.be/",
		},
		"spotify": {
			"https://open.spotify.com",
		},
	}
)

type botState struct {
	chisa       *bot.Bot
	interaction *discordgo.InteractionCreate
}

func GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
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
}

func GetCommandHandlers() func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
	return func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		bot := botState{
			chisa,
			interaction,
		}
		switch options := interaction.ApplicationCommandData().Options; options[0].Name {
		case "play":
			play(&bot, options[0].Options[0].StringValue())
		case "skip":
			skip(chisa.Session, interaction)
		case "stop":
			stop(chisa.Session, interaction)
		case "disconnect":
			disconnect(chisa.Session, interaction)
		}
	}
}
