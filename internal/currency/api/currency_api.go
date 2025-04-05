package currencyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

type currencyApiCurrencyProperties struct {
	Symbol        string   `json:"symbol"`
	Name          string   `json:"name"`
	SymbolNative  string   `json:"symbol_native"`
	DecimalDigits int      `json:"decimal_digits"`
	Rounding      float64  `json:"rounding"`
	Code          string   `json:"code"`
	NamePlural    string   `json:"name_plural"`
	Type          string   `json:"type"`
	Countries     []string `json:"countries"`
}

type currenciyApiCurrenciesResponse struct {
	Data map[string]currencyApiCurrencyProperties `json:"data"`
}

type currencyApiRateMeta struct {
	LastUpdatedAt string `json:"last_updated_at"`
}

type currencyApiRate struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type currencyApiRateResponse struct {
	Meta currencyApiRateMeta        `json:"meta"`
	Data map[string]currencyApiRate `json:"data"`
}

type CurrencyAPI struct {
	providerName string
	client       *http.Client
	token        string
	baseUrl      string
	apiVersion   string
}

func NewCurrencyApi(token string) *CurrencyAPI {
	return &CurrencyAPI{
		providerName: "currencyapi.com",
		client:       &http.Client{},
		token:        token,
		baseUrl:      "https://api.currencyapi.com",
		apiVersion:   "v3",
	}
}

func (c *CurrencyAPI) GetProviderName() string {
	return c.providerName
}

func (c *CurrencyAPI) GetCurrencies() ([]*discordgo.ApplicationCommandOptionChoice, error) {
	// Fetch currencies from API
	currencies, err := c.fetchCurrenciesFromApi()
	if err != nil {
		return nil, err
	}

	return c.transformCurrencies(currencies), nil
}

func (c *CurrencyAPI) fetchCurrenciesFromApi() (*currenciyApiCurrenciesResponse, error) {
	var response *currenciyApiCurrenciesResponse
	url := fmt.Sprintf("%s/%s/currencies?type=fiat", c.baseUrl, c.apiVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	req.Header.Add("apiKey", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return response, nil
}

func (c *CurrencyAPI) transformCurrencies(currencies *currenciyApiCurrenciesResponse) []*discordgo.ApplicationCommandOptionChoice {
	var transformedCurrencies []*discordgo.ApplicationCommandOptionChoice
	for _, currency := range currencies.Data {
		transformedCurrencies = append(transformedCurrencies, &discordgo.ApplicationCommandOptionChoice{
			Name:  fmt.Sprintf("%s (%s)", currency.Name, currency.Code),
			Value: currency.Code,
		})
	}

	return transformedCurrencies
}

func (c *CurrencyAPI) FetchCurrencyRate(source string, destination string) (*CurrencyRate, error) {
	// Fetch rate from API
	rate, err := c.fetchRateFromApi(source, destination)
	if err != nil {
		return nil, err
	}

	// Transform rate
	result, err := c.transformRate(rate, destination)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *CurrencyAPI) fetchRateFromApi(source string, destination string) (*currencyApiRateResponse, error) {
	var response *currencyApiRateResponse
	url := fmt.Sprintf("%s/%s/latest?base_currency=%s&currencies=%s", c.baseUrl, c.apiVersion, source, destination)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	req.Header.Add("apiKey", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return response, nil
}

func (c *CurrencyAPI) transformRate(rate *currencyApiRateResponse, destination string) (*CurrencyRate, error) {
	parsedTime, err := time.Parse(time.RFC3339, rate.Meta.LastUpdatedAt)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return &CurrencyRate{
		Rate:      rate.Data[destination].Value,
		UpdatedAt: parsedTime.Unix(),
	}, nil
}
