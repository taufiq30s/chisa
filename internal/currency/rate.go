package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (c *Currency) UpdateCurrencyRate(ctx context.Context) error {
	// Fetch conversion rate from API
	apiResponse, err := c.fetchLatestConversionRate(top5Currencies.BaseCurrency, top5Currencies.DestinationCurrency)
	if err != nil {
		return err
	}

	// Save conversion rate to cache
	for _, currency := range top5Currencies.DestinationCurrency {
		rate := apiResponse.Data[currency].Value
		reversedRate := 1.0 / rate

		err = c.rdb.Set(
			ctx,
			fmt.Sprintf("conversion_rate:%s%s", currency, top5Currencies.BaseCurrency),
			reversedRate,
			0).Err()
		if err != nil {
			return err
		}

		err = c.rdb.Set(
			ctx,
			fmt.Sprintf("conversion_rate:%s%s", top5Currencies.BaseCurrency, currency),
			rate,
			0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Currency) fetchConversionRate(ctx context.Context, baseCurrency string, destinationCurrency string) (float64, error) {
	// Fetch conversion rate from cache
	cacheKey := fmt.Sprintf("conversion_rate:%s%s", baseCurrency, destinationCurrency)
	rate, err := c.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		parsedRate, parseErr := strconv.ParseFloat(rate, 64)
		if parseErr != nil {
			return -1, parseErr
		}
		return parsedRate, nil
	}

	// Check when reverse currency order
	reversedCacheKey := fmt.Sprintf("conversion_rate:%s%s", destinationCurrency, baseCurrency)
	rate, err = c.rdb.Get(ctx, reversedCacheKey).Result()
	if err == nil {
		parsedRate, parseErr := strconv.ParseFloat(rate, 64)
		if parseErr != nil {
			return -1, parseErr
		}
		if parsedRate == 0 {
			return 0, nil
		}
		return 1.0 / parsedRate, nil
	}

	// Fetch conversion rate from API
	apiResponse, err := c.fetchLatestConversionRate(baseCurrency, []string{destinationCurrency})
	if err != nil {
		return -1, err
	}

	// Save conversion rate to cache
	err = c.rdb.Set(ctx, cacheKey, apiResponse.Data[destinationCurrency].Value, 0).Err()
	if err != nil {
		return -1, err
	}

	return apiResponse.Data[destinationCurrency].Value, nil
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
