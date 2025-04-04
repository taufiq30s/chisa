package currency

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	currencyapi "github.com/taufiq30s/chisa/internal/currency/api"
)

var featureName = "Chisa Currency"

type Currency struct {
	client              *http.Client
	rdb                 *redis.Client
	provider            currencyapi.CurrencyProvider
	supportedCurrencies []*discordgo.ApplicationCommandOptionChoice
}

func New(token string, rdb *redis.Client) Currency {
	return Currency{
		client:   &http.Client{},
		provider: currencyapi.NewWise(token),
		rdb:      rdb,
	}
}
