package currency

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

var baseUrl = "https://api.currencyapi.com/v3"
var featureName = "Chisa Currency"

type Currency struct {
	client              *http.Client
	token               string
	rdb                 *redis.Client
	supportedCurrencies []*discordgo.ApplicationCommandOptionChoice
}

var top5Currencies UpdateCurrencyRateDto = UpdateCurrencyRateDto{
	BaseCurrency: "IDR",
	DestinationCurrency: []string{
		"USD",
		"EUR",
		"JPY",
		"MYR",
		"SGD",
	},
}

func New(token string, rdb *redis.Client) Currency {
	return Currency{
		client: &http.Client{},
		token:  token,
		rdb:    rdb,
	}
}
