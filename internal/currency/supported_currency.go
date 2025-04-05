package currency

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"github.com/taufiq30s/chisa/utils"
)

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

	if c.provider == nil {
		fmt.Println("Currency: Provider is not set")
		utils.ErrorLog.Println("Currency: Provider is not set")
		return nil
	}
	currencies, err := c.provider.GetCurrencies()
	if err != nil {
		fmt.Printf("Error when fetching data: %v", err)
		utils.ErrorLog.Printf("Error when fetching data: %v", err)
		return nil
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
