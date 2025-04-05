package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	currencyapi "github.com/taufiq30s/chisa/internal/currency/api"
	"github.com/taufiq30s/chisa/utils"
)

var currencyRateCacheKey = "currency_rate"
var ttl_rate = 2 * time.Hour

func (c *Currency) fetchConversionRate(ctx context.Context, baseCurrency string, destinationCurrency string) (*currencyapi.CurrencyRate, error) {
	// Fetch conversion rate from cache
	cacheKey := fmt.Sprintf("%s%s", baseCurrency, destinationCurrency)
	rateDataStr, err := c.rdb.HGet(ctx, currencyRateCacheKey, cacheKey).Result()
	if err != redis.Nil {
		var rateData *currencyapi.CurrencyRate
		err := json.Unmarshal([]byte(rateDataStr), &rateData)
		if err != nil {
			utils.ErrorLog.Printf("Failed to load cache of convertion rate: %v\n", err)
			return nil, fmt.Errorf("failed to load cache")
		}
		return rateData, nil
	}

	// Fetch conversion rate from API
	if c.provider == nil {
		utils.ErrorLog.Println("Currency Provider is not set")
		fmt.Println("Currency Provider is not set")
		return nil, fmt.Errorf("currency provider is not set")
	}
	rate, err := c.provider.FetchCurrencyRate(baseCurrency, destinationCurrency)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	// Save conversion rate to cache
	rateDataMarshal, err := json.Marshal(rate)
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	err = c.rdb.HSet(ctx, currencyRateCacheKey, cacheKey, rateDataMarshal).Err()
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}
	err = c.rdb.HExpire(ctx, currencyRateCacheKey, ttl_rate, cacheKey).Err()
	if err != nil {
		utils.ErrorLog.Println(err)
		return nil, err
	}

	return rate, nil
}
