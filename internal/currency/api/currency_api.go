package currencyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
)

type CurrencyApiCurrencyProperties struct {
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

type CurrenciyApiCurrenciesResponse struct {
	Data map[string]CurrencyApiCurrencyProperties `json:"data"`
}

type CurrencyApiRateMeta struct {
	LastUpdatedAt string `json:"last_updated_at"`
}

type CurrencyApiRate struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type CurrencyApiRateResponse struct {
	Meta CurrencyApiRateMeta        `json:"meta"`
	Data map[string]CurrencyApiRate `json:"data"`
}

type CurrencyAPI struct {
	client     *http.Client
	token      string
	baseUrl    string
	apiVersion string
}

func NewCurrencyApi(token string) *CurrencyAPI {
	return &CurrencyAPI{
		client:     &http.Client{},
		token:      token,
		baseUrl:    "https://api.currencyapi.com",
		apiVersion: "v3",
	}
}

func (c *CurrencyAPI) GetCurrencies() ([]*discordgo.ApplicationCommandOptionChoice, error) {
	// Fetch currencies from API
	currencies, err := c.fetchCurrenciesFromApi()
	if err != nil {
		return nil, err
	}

	return c.transformCurrencies(currencies), nil
}

func (c *CurrencyAPI) fetchCurrenciesFromApi() (*CurrenciyApiCurrenciesResponse, error) {
	var response *CurrenciyApiCurrenciesResponse
	url := fmt.Sprintf("%s/%s/currencies?type=fiat", c.baseUrl, c.apiVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *CurrencyAPI) transformCurrencies(currencies *CurrenciyApiCurrenciesResponse) []*discordgo.ApplicationCommandOptionChoice {
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

func (c *CurrencyAPI) fetchRateFromApi(source string, destination string) (*CurrencyApiRateResponse, error) {
	var response *CurrencyApiRateResponse
	url := fmt.Sprintf("%s/%s/latest?base_currency=%s&currencies=%s", c.baseUrl, c.apiVersion, source, destination)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("apiKey", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *CurrencyAPI) transformRate(rate *CurrencyApiRateResponse, destination string) (*CurrencyRate, error) {
	parsedTime, err := time.Parse(time.RFC3339, rate.Meta.LastUpdatedAt)
	if err != nil {
		return nil, err
	}

	return &CurrencyRate{
		Rate:      rate.Data[destination].Value,
		UpdatedAt: parsedTime.Unix(),
	}, nil
}
