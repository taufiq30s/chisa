package config

import "time"

// defaults holds the fallback values for optional configuration.
const (
	DefaultRedisPoolSize              = 10
	DefaultCurrencyProvider           = "wise"
	DefaultVersion                    = "0.0.1"
	DefaultQuizWSPort                 = "8080"
	DefaultNodeConnectionTimeout      = 10 * time.Second
	DefaultSearchCacheExpiryDuration  = 1 * time.Minute
	DefaultSearchCacheCleanupInterval = 1 * time.Minute
)
