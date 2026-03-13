package currency

import (
	"bytes"
	"context"
	"encoding/gob"
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
	rdbCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(c.supportedCurrencies) > 0 {
		return c.filterResult(rdbCtx, query, c.supportedCurrencies)
	}

	rdbCache, err := c.rdb.Get(rdbCtx, "currencies").Result()
	if err != redis.Nil {
		var currencies []*discordgo.ApplicationCommandOptionChoice
		err := json.Unmarshal([]byte(rdbCache), &currencies)
		if err != nil {
			utils.ErrorLog.Printf("Error when unmarshalling data: %v", err)
			return nil
		}
		c.supportedCurrencies = currencies

		return c.filterResult(rdbCtx, query, c.supportedCurrencies)
	}

	if c.provider == nil {
		fmt.Println("Currency: Provider is not set")
		utils.ErrorLog.Println("Currency: Provider is not set")
		return nil
	}
	currencies, err := c.provider.GetCurrencies()
	if err != nil {
		utils.ErrorLog.Printf("Error when fetching data: %v", err)
		return nil
	}

	// Sort the array by the "Name" field
	sort.Slice(currencies, func(i, j int) bool {
		return currencies[i].Name < currencies[j].Name
	})

	jsonData, err := json.Marshal(currencies)
	if err != nil {
		utils.ErrorLog.Printf("Error when marshalling data: %v", err)
		return nil
	}

	// Store to redis with 1 week
	rdbCtx = context.Background()
	defer rdbCtx.Done()
	err = c.rdb.Set(rdbCtx, "currencies", string(jsonData), 7*24*time.Hour).Err()
	if err != nil {
		utils.ErrorLog.Printf("Error when storing data: %v", err)
		return nil
	}
	c.supportedCurrencies = currencies

	return c.filterResult(rdbCtx, query, currencies)
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
func (c *Currency) filterResult(ctx context.Context, query string, data []*discordgo.ApplicationCommandOptionChoice) []*discordgo.ApplicationCommandOptionChoice {
	limit := 25 // Maximum options that discord option can store
	maxLenQueryToBeCached := 2
	if query == "" {
		return data[:limit]
	}

	var result []*discordgo.ApplicationCommandOptionChoice
	var buff bytes.Buffer
	query = strings.ToLower(query)

	caches, err := c.rdb.Get(ctx, "currency:autofill:"+query).Bytes()
	if err != redis.Nil {
		buff := bytes.NewBuffer(caches)
		dec := gob.NewDecoder(buff)
		decErr := dec.Decode(&result)
		if decErr != nil {
			utils.ErrorLog.Println("Failed to dencode autofill result:", err)
			fmt.Println("Failed to dencode autofill result:", err)
		}
		return result
	}

	for _, datum := range data {
		name := strings.ToLower(datum.Name)
		if strings.Contains(name, query) {
			result = append(result, datum)
		}
	}

	// Convert result into gob and cache it into redis
	if len(result) > limit {
		result = result[:limit]
	}
	if len(query) < maxLenQueryToBeCached {
		return result
	}
	enc := gob.NewEncoder(&buff)
	err = enc.Encode(result)
	if err != nil {
		utils.ErrorLog.Println("Failed to encode autofill result:", err)
		fmt.Println("Failed to encode autofill result:", err)
	} else {
		err = c.rdb.SetEx(ctx, "currency:autofill:"+query, buff.Bytes(), time.Hour).Err()
		if err != nil {
			utils.ErrorLog.Println("Failed to store cache:", err)
			fmt.Println("Failed to store cache:", err)
		}
	}

	return result
}
