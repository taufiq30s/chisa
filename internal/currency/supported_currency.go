package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

type CurrencyProperties struct {
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

type CurrenciesApiResponse struct {
	Data map[string]CurrencyProperties `json:"data"`
}

// GetCurrencies retrieves a list of supported currencies and returns them as Discord command options
//
// Parameters:
//   - query: Search string to filter currencies by name
//
// Returns:
//   - Array of Discord ApplicationCommandOptionChoice containing filtered currency options
//   - Maximum 25 results are returned to comply with Discord limits
//   - Each option contains currency name and code (e.g. "US Dollar (USD)") and have code as value
func (c *Currency) GetCurrencies(query string) []*discordgo.ApplicationCommandOptionChoice {
	if len(c.supportedCurrencies) > 0 {
		return c.filterResult(query, c.supportedCurrencies)
	}

	rdbCtx := context.Background()
	defer rdbCtx.Done()
	rdbCache, err := c.rdb.Get(rdbCtx, "currencies").Result()
	if err != redis.Nil {
		var currencies []*discordgo.ApplicationCommandOptionChoice
		err := json.Unmarshal([]byte(rdbCache), &currencies)
		if err != nil {
			fmt.Printf("Error when unmarshalling data: %v", err)
			utils.ErrorLog.Printf("Error when unmarshalling data: %v", err)
			return nil
		}
		c.supportedCurrencies = currencies

		return c.filterResult(query, c.supportedCurrencies)
	}

	apiData, err := c.fetchCurrenciesFromAPI()
	if err != nil {
		fmt.Printf("Error when fetching data: %v", err)
		utils.ErrorLog.Printf("Error when fetching data: %v", err)
		return nil
	}

	var currencies []*discordgo.ApplicationCommandOptionChoice
	for _, value := range apiData.Data {
		currencies = append(currencies, &discordgo.ApplicationCommandOptionChoice{
			Name:  fmt.Sprintf("%s (%s)", value.Name, value.Code),
			Value: value.Code,
		})
	}

	// Sort the array by the "Name" field
	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Name < currencies[j].Name
	})

	jsonData, err := json.Marshal(currencies)
	if err != nil {
		fmt.Printf("Error when marshalling data: %v", err)
		utils.ErrorLog.Printf("Error when marshalling data: %v", err)
		return nil
	}

	// Store to redis with 1 week
	rdbCtx = context.Background()
	defer rdbCtx.Done()
	err = c.rdb.Set(rdbCtx, "currencies", string(jsonData), 7*24*time.Hour).Err()
	if err != nil {
		fmt.Printf("Error when storing data: %v", err)
		utils.ErrorLog.Printf("Error when storing data: %v", err)
		return nil
	}
	c.supportedCurrencies = currencies
	return c.filterResult(query, currencies)
}

// filterResult filters and limits currency options based on a search query
//
// Parameters:
//   - query: Search string to filter currencies by name (case-insensitive)
//   - data: Array of currency options to filter
//
// Returns:
//   - Filtered array of currency options that match the search query
//   - If query is empty, returns first 25 items from data
//   - If query is provided, returns up to 25 matching items
//   - Items are matched if their name contains the query string (case-insensitive)
func (c *Currency) filterResult(query string, data []*discordgo.ApplicationCommandOptionChoice) []*discordgo.ApplicationCommandOptionChoice {
	limit := 25 // Maximum options that discord option can store
	if query == "" {
		return data[:limit]
	}

	var result []*discordgo.ApplicationCommandOptionChoice
	query = strings.ToLower(query)

	for _, datum := range data {
		name := strings.ToLower(datum.Name)
		if strings.Contains(name, query) {
			result = append(result, datum)
		}
	}
	if len(result) < limit {
		return result
	}
	return result[:limit]
}

// fetchCurrenciesFromAPI retrieves currency data from the external API
//
// Makes an HTTP GET request to fetch supported fiat currencies from the API endpoint
// Requires an API key token for authentication
//
// Returns:
//   - CurrenciesApiResponse containing map of currency codes to their properties
//   - Error if the request fails, returns non-200 status, or response parsing fails
//
// The response includes details like:
//   - Currency name, code and symbol
//   - Native symbol and plural name
//   - Decimal digits and rounding
//   - List of countries where used
func (c *Currency) fetchCurrenciesFromAPI() (*CurrenciesApiResponse, error) {
	var apiResponse *CurrenciesApiResponse
	url := fmt.Sprintf("%s/currencies?type=fiat", baseUrl)

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

	// Store to apiResponse
	err = json.Unmarshal(resBody, &apiResponse)
	if err != nil {
		return nil, err
	}

	// Get data from apiResponse
	return apiResponse, nil
}
