package currency

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/internal/responses"
	"github.com/taufiq30s/chisa/utils"
)

type ConversionDto struct {
	Amount              float64
	BaseCurrency        string
	DestinationCurrency string
}

type conversionResultDto struct {
	Result    float64
	Rate      float64
	UpdatedAt string
}

func (c *Currency) Conversion(s *discordgo.Session, i *discordgo.InteractionCreate, data *ConversionDto) {
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
		}).Execute()
	}

	// Make sure 'from' and 'to' in uppercase
	data.BaseCurrency = strings.ToUpper(data.BaseCurrency)
	data.DestinationCurrency = strings.ToUpper(data.DestinationCurrency)

	// calculate result and return value
	result, err := c.calculateCurrency(data.Amount, data.BaseCurrency, data.DestinationCurrency)
	if err != nil {
		responses.ErrorResponse(s, i, &responses.ErrorResponseData{
			Feature:     featureName,
			Title:       "Conversion Failed",
			Description: "Please try again later or contact Pengasuh Anak",
		}).ExecuteDefer()
		return
	}

	// Show result to user
	var strRate string
	if result.Rate < RateDisplayThreshold {
		strRate = fmt.Sprintf("%.5f", result.Rate)
	} else {
		strRate = utils.FormatDecimalNumberWithGrouping(result.Rate)
	}
	body := responses.CreateMessageEmbed(
		s, fmt.Sprintf("Exchange from %s to %s", data.BaseCurrency, data.DestinationCurrency),
		fmt.Sprintf(
			"**%s %s** = %s %s",
			utils.FormatDecimalNumberWithGrouping(data.Amount),
			data.BaseCurrency,
			utils.FormatDecimalNumberWithGrouping(result.Result),
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
				Inline: true,
			},
			{
				Name:   "Last updated",
				Value:  result.UpdatedAt,
				Inline: true,
			},
			{
				Name:  "Powered by",
				Value: c.provider.GetProviderName(),
			},
		}),
		responses.SetColor("42F30B"),
	)
	responses.InteractionResponse(s, i.Interaction).WithEmbed(body).SendDefer()
}

func (c *Currency) validateConversionInput(amount float64, baseCurrency string, destinationCurrency string) error {
	if amount <= MinimumAmount {
		return fmt.Errorf("value must be greater than %d", MinimumAmount)
	}
	if baseCurrency == destinationCurrency {
		return fmt.Errorf("base currency and target currency must be different")
	}
	if baseCurrency == "" || destinationCurrency == "" {
		return fmt.Errorf("base currency and target currency must be provided")
	}
	if len(baseCurrency) != CurrencyCodeLength || len(destinationCurrency) != CurrencyCodeLength {
		return fmt.Errorf("base currency and target currency must be %d characters", CurrencyCodeLength)
	}
	return nil
}

func (c *Currency) calculateCurrency(amount float64, baseCurrency string, destinationCurrency string) (*conversionResultDto, error) {
	// Fetch conversion rate
	rateData, err := c.fetchConversionRate(context.Background(), baseCurrency, destinationCurrency)
	if err != nil {
		return nil, err
	}
	// Calculate conversion
	return &conversionResultDto{
		Result:    amount * rateData.Rate,
		Rate:      rateData.Rate,
		UpdatedAt: time.Unix(rateData.UpdatedAt, MinimumAmount).Format(DateFormat),
	}, nil
}
