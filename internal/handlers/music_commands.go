package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
)

var (
	musicCommandHandler = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[0].Name {
		case "play":
			chisa.Music.Play(chisa.Session, interaction, options[0].Options[0].StringValue())
		case "skip":
			// skip(chisa.Session, interaction)
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
)
