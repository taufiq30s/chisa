package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/currency"
	"github.com/taufiq30s/chisa/utils"
)

var (
	currencyCommandOptions = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[0].Name {
		case "convert":
			var choices []*discordgo.ApplicationCommandOptionChoice

			for _, option := range interaction.ApplicationCommandData().Options[0].Options {
				if option.Focused {
					choices = chisa.Currency.GetCurrencies(option.StringValue())
					break
				}
			}
			err := chisa.Session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices,
				},
			})
			if err != nil {
				fmt.Println("Error when send options :", err)
				utils.ErrorLog.Println("Error when send options :", err)
			}
		}
	}
	currencyCommandHandler = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[0].Name {
		case "convert":
			switch interaction.Type {
			case discordgo.InteractionApplicationCommand:
				data := interaction.ApplicationCommandData()
				chisa.Currency.Conversion(chisa.Session, interaction, &currency.ConversionDto{
					Amount:              data.Options[0].Options[0].FloatValue(),
					BaseCurrency:        data.Options[0].Options[1].StringValue(),
					DestinationCurrency: data.Options[0].Options[2].StringValue(),
				})
			}
		}
	}
	currencyCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "currency",
			Description: "Currency commands",
			Type:        discordgo.ChatApplicationCommand,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "convert",
					Description: "Convert currency",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "amount",
							Description: "Amount to convert",
							Type:        discordgo.ApplicationCommandOptionNumber,
							Required:    true,
						},
						{
							Name:         "from",
							Description:  "Currency to convert from",
							Type:         discordgo.ApplicationCommandOptionString,
							Required:     true,
							Autocomplete: true,
						},
						{
							Name:         "to",
							Description:  "Currency to convert to",
							Type:         discordgo.ApplicationCommandOptionString,
							Required:     true,
							Autocomplete: true,
						},
					},
				},
			},
		},
	}
)
