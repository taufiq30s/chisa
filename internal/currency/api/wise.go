package currencyapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/taufiq30s/chisa/utils"
)

type WiseCurrencyProperties struct {
	Code             string   `json:"code"`
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	CountryKeywords  []string `json:"countryKeywords"`
	SupportsDecimals bool     `json:"supportsDecimals"`
}

type WiseRateResponse struct {
	Rate   float64 `json:"rate"`
	Source string  `json:"source"`
	Target string  `json:"target"`
	Time   string  `json:"time"`
}

type Wise struct {
	providerName string
	client       *http.Client
	token        string
	baseUrl      string
	apiVersion   string
}

func NewWise(token string) *Wise {
	return &Wise{
		providerName: "Wise",
		client:       &http.Client{},
		token:        token,
		baseUrl:      "https://api.wise.com",
		apiVersion:   "v1",
	}
}

func (w *Wise) GetProviderName() string {
	return w.providerName
}

func (w *Wise) GetCurrencies() ([]*discordgo.ApplicationCommandOptionChoice, error) {
	wiseCurrencies, err := w.fetchCurrenciesFromApi()
	if err != nil {
		return nil, err
	}

	return w.transformCurrencies(wiseCurrencies), nil
}

func (w *Wise) fetchCurrenciesFromApi() ([]*WiseCurrencyProperties, error) {
	response := make([]*WiseCurrencyProperties, 0)
	url := fmt.Sprintf("%s/%s/currencies?type=fiat", w.baseUrl, w.apiVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.token))

	resp, err := w.client.Do(req)
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

func (w *Wise) transformCurrencies(currencies []*WiseCurrencyProperties) []*discordgo.ApplicationCommandOptionChoice {
	var transformedCurrencies []*discordgo.ApplicationCommandOptionChoice
	for _, currency := range currencies {
		transformedCurrencies = append(transformedCurrencies, &discordgo.ApplicationCommandOptionChoice{
			Name:  fmt.Sprintf("%s (%s)", currency.Name, currency.Code),
			Value: currency.Code,
		})
	}

	return transformedCurrencies
}

func (w *Wise) FetchCurrencyRate(source string, destination string) (*CurrencyRate, error) {
	wiseRate, err := w.fetchRateFromApi(source, destination)
	if err != nil {
		return nil, err
	}

	return w.transformRate(wiseRate)
}

func (w *Wise) fetchRateFromApi(source string, destination string) (*WiseRateResponse, error) {
	var response []*WiseRateResponse
	url := fmt.Sprintf("%s/%s/rates?source=%s&target=%s", w.baseUrl, w.apiVersion, source, destination)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.token))

	resp, err := w.client.Do(req)
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

	return response[0], nil
}

func (w *Wise) transformRate(rate *WiseRateResponse) (*CurrencyRate, error) {
	parsedTime, err := time.Parse("2006-01-02T15:04:05-0700", rate.Time)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return &CurrencyRate{
		Rate:      rate.Rate,
		UpdatedAt: parsedTime.Unix(),
	}, nil
}

type wiseQuote struct {
	Rate           float64 `json:"rate"`
	RateTimestamp  string  `json:"rateTimestamp"`
	PaymentOptions []struct {
		SourceCurrency string  `json:"sourceCurrency"`
		TargetCurrency string  `json:"targetCurrency"`
		SourceAmount   float64 `json:"sourceAmount"`
		TargetAmount   float64 `json:"targetAmount"`
		Fee            struct {
			Total float64 `json:"total"`
		} `json:"fee"`
	} `json:"paymentOptions"`
}

type wiseQuoteBody struct {
	SourceCurrency string  `json:"sourceCurrency"`
	TargetCurrency string  `json:"targetCurrency"`
	SourceAmount   float64 `json:"sourceAmount,omitempty"`
	TargetAmount   float64 `json:"targetAmount,omitempty"`
}

type WiseSimulate struct {
	Amount        float64
	WiseFee       float64
	Total         float64
	ReceivedTotal float64
	Rate          float64
	UpdatedAt     int64
}

func SimulateWiseTransfer(amount float64, baseCurrency string, destinationCurrency string, isSourceAmount bool) (*WiseSimulate, error) {
	var response *wiseQuote
	var baseUrl = "https://api.wise.com"
	var client = &http.Client{}

	url := fmt.Sprintf("%s/v3/quotes/", baseUrl)

	// Prepare body
	body := createQuoteBody(amount, baseCurrency, destinationCurrency, isSourceAmount)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	// Prepare request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		utils.ErrorLog.Printf("failed to simulate wise quote: %s\n", resp.Status)
		return nil, fmt.Errorf("failed to simulate wise quote: %s", resp.Status)
	}

	// Extract response body
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	// Formating output
	parserTime, err := time.Parse(time.RFC3339, response.RateTimestamp)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return &WiseSimulate{
		Amount:        response.PaymentOptions[0].SourceAmount,
		WiseFee:       float64(response.PaymentOptions[0].Fee.Total),
		Total:         response.PaymentOptions[0].SourceAmount,
		ReceivedTotal: response.PaymentOptions[0].TargetAmount,
		Rate:          1.0 / response.Rate,
		UpdatedAt:     parserTime.Unix(),
	}, nil
}

func createQuoteBody(amount float64, baseCurrency string, destinationCurrency string, isSourceAmount bool) *wiseQuoteBody {
	body := &wiseQuoteBody{
		SourceCurrency: baseCurrency,
		TargetCurrency: destinationCurrency,
	}
	if isSourceAmount {
		body.SourceAmount = amount
	} else {
		body.TargetAmount = amount
	}
	return body
}
