package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/taufiq30s/chisa/utils"
)

type ConversionMeta struct {
	LastUpdatedAt string `json:"last_updated_at"`
}

type ConversionRate struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
}

type ConversionApiResponse struct {
	Meta ConversionMeta            `json:"meta"`
	Data map[string]ConversionRate `json:"data"`
}

type UpdateCurrencyRateDto struct {
	BaseCurrency        string
	DestinationCurrency []string
}

type CurrencyRateDto struct {
	Rate      float64 `json:"rate"`
	UpdatedAt int64   `json:"updated_at"`
}

var currencyRateCacheKey = "currency_rate"

func (c *Currency) UpdateCurrencyRate(ctx context.Context) error {
	// Fetch conversion rate from API
	apiResponse, err := c.fetchLatestConversionRate(top5Currencies.BaseCurrency, top5Currencies.DestinationCurrency)
	if err != nil {
		return err
	}

	// TODO: Delete all outdate currency_rate

	// Save conversion rate to cache
	hashData := make(map[string]any)
	for _, currency := range top5Currencies.DestinationCurrency {
		rate := apiResponse.Data[currency].Value
		reversedRate := 1.0 / rate
		updatedStr := apiResponse.Meta.LastUpdatedAt
		parsedTime, err := time.Parse(time.RFC3339, updatedStr)
		if err != nil {
			return err
		}

		rateData, err := json.Marshal(CurrencyRateDto{
			Rate:      rate,
			UpdatedAt: parsedTime.Unix(),
		})
		if err != nil {
			return err
		}

		reversedRateData, err := json.Marshal(CurrencyRateDto{
			Rate:      reversedRate,
			UpdatedAt: parsedTime.Unix(),
		})
		if err != nil {
			return err
		}

		hashData[fmt.Sprintf("%s%s", currency, top5Currencies.BaseCurrency)] = reversedRateData
		hashData[fmt.Sprintf("%s%s", top5Currencies.BaseCurrency, currency)] = rateData
	}

	err = c.rdb.HSet(ctx, currencyRateCacheKey, hashData).Err()
	if err != nil {
		return err
	}
	fmt.Println("Currency rate updated")
	utils.InfoLog.Println("Currency rate updated")
	return nil
}

func (c *Currency) fetchConversionRate(ctx context.Context, baseCurrency string, destinationCurrency string) (*CurrencyRateDto, error) {
	// Fetch conversion rate from cache
	cacheKey := fmt.Sprintf("%s%s", baseCurrency, destinationCurrency)
	rateDataStr, err := c.rdb.HGet(ctx, currencyRateCacheKey, cacheKey).Result()
	if err == nil {
		var rateData *CurrencyRateDto
		err := json.Unmarshal([]byte(rateDataStr), &rateData)
		if err != nil {
			utils.ErrorLog.Printf("Failed to load cache of convertion rate: %v\n", err)
			return nil, fmt.Errorf("failed to load cache")
		}
		return rateData, nil
	}

	// Fetch conversion rate from API
	apiResponse, err := c.fetchLatestConversionRate(baseCurrency, []string{destinationCurrency})
	if err != nil {
		return nil, err
	}

	// Save conversion rate to cache
	rateData := &CurrencyRateDto{
		Rate:      apiResponse.Data[destinationCurrency].Value,
		UpdatedAt: time.Now().Unix(),
	}
	rateDataMarshal, err := json.Marshal(rateData)
	if err != nil {
		return nil, err
	}

	err = c.rdb.HSet(ctx, currencyRateCacheKey, cacheKey, rateDataMarshal).Err()
	if err != nil {
		return nil, err
	}

	return rateData, nil
}

func (c *Currency) fetchLatestConversionRate(baseCurrency string, destinationCurrency []string) (*ConversionApiResponse, error) {
	var apiResponse *ConversionApiResponse
	url := fmt.Sprintf("%s/latest?type=fiat&base_currency=%s&currencies=%s", baseUrl, baseCurrency, strings.Join(destinationCurrency, ","))

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("apikey", c.token)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error when doing request: %v", resp.StatusCode)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(resBody, &apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse, nil
}
