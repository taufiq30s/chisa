package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/bot"
	"github.com/taufiq30s/chisa/internal/currency"
	"github.com/taufiq30s/chisa/utils"
)

var (
	currencyCommandOptions = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[OptionIndexFirst].Name {
		case "convert", "simulate":
			var choices []*discordgo.ApplicationCommandOptionChoice

			for _, option := range interaction.ApplicationCommandData().Options[OptionIndexFirst].Options {
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
				utils.ErrorLog.Println("Error when send options :", err)
			}
		}
	}
	currencyCommandHandler = func(chisa *bot.Bot, interaction *discordgo.InteractionCreate) {
		switch options := interaction.ApplicationCommandData().Options; options[OptionIndexFirst].Name {
		case "convert":
			data := interaction.ApplicationCommandData()
			chisa.Currency.Conversion(chisa.Session, interaction, &currency.ConversionDto{
				Amount:              data.Options[OptionIndexFirst].Options[OptionIndexFirst].FloatValue(),
				BaseCurrency:        data.Options[OptionIndexFirst].Options[OptionIndexSecond].StringValue(),
				DestinationCurrency: data.Options[OptionIndexFirst].Options[OptionIndexThird].StringValue(),
			})
		case "simulate":
			data := interaction.ApplicationCommandData()
			isSourceAmount := true
			if data.Options[OptionIndexFirst].Options[OptionIndexFirst].StringValue() == "receiver" {
				isSourceAmount = false
			}
			chisa.Currency.SimulateWise(chisa.Session, interaction, &currency.WiseSimulateDto{
				Amount:              data.Options[OptionIndexFirst].Options[OptionIndexSecond].FloatValue(),
				BaseCurrency:        data.Options[OptionIndexFirst].Options[OptionIndexThird].StringValue(),
				DestinationCurrency: data.Options[OptionIndexFirst].Options[OptionIndexFourth].StringValue(),
				IsSourceAmount:      isSourceAmount,
			})
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
				{
					Name:        "simulate",
					Description: "Simulate Wise Transfer",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "mode",
							Description: "Mode of amount's currency.",
							Type:        discordgo.ApplicationCommandOptionString,
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "Using Source Currency",
									Value: "send",
								},
								{
									Name:  "Using Target Currency",
									Value: "receiver",
								},
							},
						},
						{
							Name:        "amount",
							Description: "Amount to simulate",
							Type:        discordgo.ApplicationCommandOptionNumber,
							Required:    true,
						},
						{
							Name:         "source",
							Description:  "Source Currency of Sender",
							Type:         discordgo.ApplicationCommandOptionString,
							Required:     true,
							Autocomplete: true,
						},
						{
							Name:         "target",
							Description:  "Target Currency of Receiver",
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
