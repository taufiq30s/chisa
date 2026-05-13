package currency

import (
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/internal/config"
	currencyapi "github.com/taufiq30s/chisa/internal/currency/api"
	"github.com/taufiq30s/chisa/utils"
)

const featureName = "Chisa Currency"
const (
	CurrencyApiProvider string = "currency_api"
	WiseProvider        string = "wise"
)

// CurrencyService defines the external interface for the currency feature.
// Handlers interact with Currency only through this interface.
type CurrencyService interface {
	Conversion(s *discordgo.Session, i *discordgo.InteractionCreate, data *ConversionDto)
	SimulateWise(s *discordgo.Session, i *discordgo.InteractionCreate, data *WiseSimulateDto)
	GetCurrencies(query string) []*discordgo.ApplicationCommandOptionChoice
}

type Currency struct {
	client              *http.Client
	rdb                 *redis.Client
	provider            currencyapi.CurrencyProvider
	supportedCurrencies []*discordgo.ApplicationCommandOptionChoice
}

func New(provider string, cfg *config.Config, rdb *redis.Client) *Currency {
	currency := &Currency{
		client: &http.Client{},
		rdb:    rdb,
	}
	switch provider {
	case "currency_api":
		currency.provider = currencyapi.NewCurrencyApi(cfg.CurrencyAPIToken)
	case "wise":
		currency.provider = currencyapi.NewWise(cfg.WiseToken)
	default:
		utils.ErrorLog.Printf("Invalid Currency Provider \"%s\"\n", provider)
	}
	return currency
}
