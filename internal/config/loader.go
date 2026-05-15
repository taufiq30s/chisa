package config

import (
	"fmt"
	"os"
	"strings"
)

// Load reads all required and optional environment variables, validates them,
// and returns a populated *Config.  It fails fast with a descriptive error
// listing every missing required variable so operators can fix everything in
// one restart.
func Load() (*Config, error) {
	var missing []string

	get := func(key string) string {
		v, ok := os.LookupEnv(key)
		if !ok || v == "" {
			missing = append(missing, key)
		}
		return v
	}

	getOptional := func(key, fallback string) string {
		v, ok := os.LookupEnv(key)
		if !ok || v == "" {
			return fallback
		}
		return v
	}

	cfg := &Config{
		// Required
		BotToken: get("BOT_TOKEN"),
		GuildID:  get("AKASHIC_SERVER_ID"),
		RedisURL: get("REDIS_URL"),

		// Moderation IDs (required)
		VerifiedRoleID:   get("AKASHIC_VERIFIED_ROLE_ID"),
		VerifChannelID:   get("AKASHIC_VERIF_CHANNEL_ID"),
		ModChannelID:     get("CHISA_MOD_CHANNEL_ID"),
		LogChannelID:     get("CHISA_LOG_CHANNEL_ID"),
		WelcomeChannelID: get("AKASHIC_WELCOME_CHANNEL_ID"),
		RulesChannelID:   get("AKASHIC_RULE_CHANNEL_ID"),

		// Optional with defaults
		CurrencyProvider: getOptional("CURRENCY_PROVIDER", DefaultCurrencyProvider),
		WiseToken:        getOptional("WISE_TOKEN", ""),
		CurrencyAPIToken: getOptional("CURRENCY_API_TOKEN", ""),
		Version:          getOptional("VERSION", DefaultVersion),
		QuizWSPort:       getOptional("QUIZ_WS_PORT", DefaultQuizWSPort),
		PoolSize:         DefaultRedisPoolSize,

		// Timing constants (not from env — set from package defaults)
		NodeConnectionTimeout:      DefaultNodeConnectionTimeout,
		SearchCacheExpiryDuration:  DefaultSearchCacheExpiryDuration,
		SearchCacheCleanupInterval: DefaultSearchCacheCleanupInterval,
	}

	if err := validate(cfg, missing); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validate(cfg *Config, missing []string) error {
	if len(missing) > 0 {
		return fmt.Errorf(
			"missing required environment variables: %s\n"+
				"Copy .env.example to .env and fill in the values",
			strings.Join(missing, ", "),
		)
	}

	switch cfg.CurrencyProvider {
	case "wise":
		if cfg.WiseToken == "" {
			return fmt.Errorf("WISE_TOKEN is required when CURRENCY_PROVIDER=wise")
		}
	case "currency_api":
		if cfg.CurrencyAPIToken == "" {
			return fmt.Errorf("CURRENCY_API_TOKEN is required when CURRENCY_PROVIDER=currency_api")
		}
	default:
		return fmt.Errorf("invalid CURRENCY_PROVIDER %q: must be \"wise\" or \"currency_api\"", cfg.CurrencyProvider)
	}

	return nil
}
