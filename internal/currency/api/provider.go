package currencyapi

import "github.com/bwmarrin/discordgo"

type CurrencyRate struct {
	Rate      float64 `json:"rate"`
	UpdatedAt int64   `json:"updated_at"`
}

type CurrencyProvider interface {
	GetCurrencies() ([]*discordgo.ApplicationCommandOptionChoice, error)
	FetchCurrencyRate(source string, destination string) (*CurrencyRate, error)
}
