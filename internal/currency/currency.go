package currency

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	currencyapi "github.com/taufiq30s/chisa/internal/currency/api"
	"github.com/taufiq30s/chisa/utils"
)

const featureName = "Chisa Currency"
const (
	CurrencyApiProvider string = "currency_api"
	WiseProvider        string = "wise"
)

type Currency struct {
	client              *http.Client
	rdb                 *redis.Client
	provider            currencyapi.CurrencyProvider
	supportedCurrencies []*discordgo.ApplicationCommandOptionChoice
}

func getToken(provider string) string {
	token, err := utils.GetEnv(provider)
	if err != nil {
		fmt.Println("Failed to get token :", err)
		utils.ErrorLog.Fatalln("Failed to get token :", err)
	}
	return token
}

func New(provider string, rdb *redis.Client) Currency {
	currency := Currency{
		client: &http.Client{},
		rdb:    rdb,
	}
	switch provider {
	case "currency_api":
		token := getToken("CURRENCY_API_TOKEN")
		currency.provider = currencyapi.NewCurrencyApi(token)
	case "wise":
		token := getToken("WISE_TOKEN")
		currency.provider = currencyapi.NewWise(token)
	default:
		utils.ErrorLog.Printf("Invalid Currency Provider \"%s\"\n", provider)
		fmt.Printf("Invalid Currency Provider \"%s\"\n", provider)
	}
	return currency
}
