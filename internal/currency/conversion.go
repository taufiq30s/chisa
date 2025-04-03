package currency

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type ConversionDto struct {
	Amount              float64
	BaseCurrency        string
	DestinationCurrency string
}

func (c *Currency) Conversion(s *discordgo.Session, i *discordgo.InteractionCreate, data *ConversionDto) {
	err := responses.InteractionResponse(s, i.Interaction).Defer()
	if err != nil {
		fmt.Println("Error when defer :", err)
		utils.ErrorLog.Println("Error when defer :", err)
	}
	// Validate input
	err = c.validateConversionInput(data.Amount, data.BaseCurrency, data.DestinationCurrency)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature: featureName,
			Title:   "Validation Failed",
			Err:     err,
		}).Execute()
	}

	// Make sure 'from' and 'to' in uppercase
	data.BaseCurrency = strings.ToUpper(data.BaseCurrency)
	data.DestinationCurrency = strings.ToUpper(data.DestinationCurrency)

	// calculate result and return value
	result, rate, err := c.calculateCurrency(data.Amount, data.BaseCurrency, data.DestinationCurrency)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature: featureName,
			Title:   "Conversion Failed",
			Err:     err,
		}).Execute()
	}

	// Show result to user
	var strRate string
	if rate < 1 {
		strRate = fmt.Sprintf("%.5f", rate)
	} else {
		strRate = utils.FormatNumberWithGrouping(rate)
	}
	body := responses.CreateMessageEmbed(
		s, fmt.Sprintf("Exchange from %s to %s", data.BaseCurrency, data.DestinationCurrency),
		fmt.Sprintf(
			"**%s %s** = %s %s",
			utils.FormatNumberWithGrouping(data.Amount),
			data.BaseCurrency,
			utils.FormatNumberWithGrouping(result),
			data.DestinationCurrency),
		featureName,
		responses.SetFields([]*discordgo.MessageEmbedField{
			{
				Name: "Exchange rate",
				Value: fmt.Sprintf(
					"1 %s = %s %s",
					data.BaseCurrency,
					strRate,
					data.DestinationCurrency,
				),
			},
		}),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(body).SendDefer()
}

func (c *Currency) validateConversionInput(amount float64, baseCurrency string, destinationCurrency string) error {
	if amount <= 0 {
		return fmt.Errorf("value must be greater than 0")
	}
	if baseCurrency == destinationCurrency {
		return fmt.Errorf("base currency and target currency must be different")
	}
	if baseCurrency == "" || destinationCurrency == "" {
		return fmt.Errorf("base currency and target currency must be provided")
	}
	if len(baseCurrency) != 3 || len(destinationCurrency) != 3 {
		return fmt.Errorf("base currency and target currency must be 3 characters")
	}
	return nil
}

func (c *Currency) calculateCurrency(amount float64, baseCurrency string, destinationCurrency string) (float64, float64, error) {
	// Fetch conversion rate
	rate, err := c.fetchConversionRate(context.Background(), baseCurrency, destinationCurrency)
	if err != nil {
		return -1, -1, err
	}

	// Calculate conversion
	return amount * rate, rate, nil
}
