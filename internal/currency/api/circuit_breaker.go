package currencyapi

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/sony/gobreaker"
	"github.com/taufiq30s/chisa/utils"
)

// circuitBreakerProvider wraps any CurrencyProvider with a circuit breaker.
// If the underlying provider fails consecutively, the breaker opens and
// fast-fails calls until a cooldown period has elapsed.
type circuitBreakerProvider struct {
	inner   CurrencyProvider
	breaker *gobreaker.CircuitBreaker
}

// NewCircuitBreakerProvider wraps provider with a circuit breaker.
// The breaker opens after 5 consecutive failures and resets after 30 s.
func NewCircuitBreakerProvider(provider CurrencyProvider) CurrencyProvider {
	settings := gobreaker.Settings{
		Name:        fmt.Sprintf("currency-provider-%s", provider.GetProviderName()),
		MaxRequests: 1,               // allow 1 probe request in half-open state
		Interval:    60 * time.Second, // reset counts every 60 s in closed state
		Timeout:     30 * time.Second, // stay open for 30 s before trying again
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			utils.WarningLog.Printf("Currency circuit breaker [%s] state: %s → %s\n", name, from, to)
		},
	}
	return &circuitBreakerProvider{
		inner:   provider,
		breaker: gobreaker.NewCircuitBreaker(settings),
	}
}

func (c *circuitBreakerProvider) GetProviderName() string {
	return c.inner.GetProviderName()
}

func (c *circuitBreakerProvider) GetCurrencies() ([]*discordgo.ApplicationCommandOptionChoice, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.inner.GetCurrencies()
	})
	if err != nil {
		return nil, err
	}
	return result.([]*discordgo.ApplicationCommandOptionChoice), nil
}

func (c *circuitBreakerProvider) FetchCurrencyRate(source, destination string) (*CurrencyRate, error) {
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.inner.FetchCurrencyRate(source, destination)
	})
	if err != nil {
		return nil, fmt.Errorf("currency provider unavailable: %w", err)
	}
	return result.(*CurrencyRate), nil
}
