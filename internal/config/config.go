// Package config provides centralised application configuration loaded from
// environment variables.  All subsystems should receive a *Config at
// construction time instead of calling utils.GetEnv themselves.
package config

import "time"

// Config holds every runtime-configurable value for the bot.
// Fields are grouped by subsystem.
type Config struct {
	// Discord
	BotToken string
	GuildID  string

	// Redis
	RedisURL string
	PoolSize int

	// Currency
	CurrencyProvider string // "wise" or "currency_api"
	WiseToken        string
	CurrencyAPIToken string

	// Moderation channel / role IDs
	VerifiedRoleID   string
	VerifChannelID   string
	ModChannelID     string
	LogChannelID     string
	WelcomeChannelID string
	RulesChannelID   string

	// App
	Version     string
	QuizWSPort  string

	// Derived / computed at load time (not from env directly)
	NodeConnectionTimeout      time.Duration
	SearchCacheExpiryDuration  time.Duration
	SearchCacheCleanupInterval time.Duration
}
