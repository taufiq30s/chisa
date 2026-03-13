package currency

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	currencyapi "github.com/taufiq30s/chisa/internal/currency/api"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type WiseSimulateDto struct {
	Amount              float64
	BaseCurrency        string
	DestinationCurrency string
	IsSourceAmount      bool
}

func (c *Currency) SimulateWise(s *discordgo.Session, i *discordgo.InteractionCreate, data *WiseSimulateDto) {
	err := responses.InteractionResponse(s, i.Interaction).Defer()
	if err != nil {
		utils.ErrorLog.Println("Error when defer :", err)
	}
	// Validate input
	err = c.validateConversionInput(data.Amount, data.BaseCurrency, data.DestinationCurrency)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature: featureName,
			Title:   "Validation Failed",
			Err:     err,
		}).ExecuteDefer()
	}

	// Make sure 'from' and 'to' in uppercase
	data.BaseCurrency = strings.ToUpper(data.BaseCurrency)
	data.DestinationCurrency = strings.ToUpper(data.DestinationCurrency)

	// Get Simulation
	wiseSimulate, err := currencyapi.SimulateWiseTransfer(data.Amount, data.BaseCurrency, data.DestinationCurrency, data.IsSourceAmount)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature: featureName,
			Title:   "Failed to get simulation",
			Err:     err,
		}).ExecuteDefer()
		return
	}

	var strRate string
	if wiseSimulate.Rate < 1 {
		strRate = fmt.Sprintf("%.5f", wiseSimulate.Rate)
	} else {
		strRate = utils.FormatDecimalNumberWithGrouping(wiseSimulate.Rate)
	}
	body := responses.CreateMessageEmbed(
		s, "Wise Send Money Simulation",
		"**Note:\n- This is simulation, not actual transaction. \n- Fee doesn't include VAT.**",
		featureName,
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name: "You send",
				Value: fmt.Sprintf(
					"%s %s",
					utils.FormatDecimalNumberWithGrouping(wiseSimulate.Total),
					data.BaseCurrency,
				),
				Inline: true,
			},
			{
				Name: "They receive",
				Value: fmt.Sprintf(
					"%s %s",
					utils.FormatDecimalNumberWithGrouping(wiseSimulate.ReceivedTotal),
					data.DestinationCurrency,
				),
				Inline: true,
			},
			{
				Inline: false,
			},
			{
				Name: "Rate",
				Value: fmt.Sprintf(
					"1 %s = %s %s",
					data.DestinationCurrency,
					strRate,
					data.BaseCurrency,
				),
				Inline: true,
			},
			{
				Name: "Fee",
				Value: fmt.Sprintf(
					"%s %s",
					utils.FormatDecimalNumberWithGrouping(wiseSimulate.WiseFee),
					data.BaseCurrency,
				),
				Inline: true,
			},
		}),
		responses.SetColor("42F30B"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(body).SendDefer()
}
