package bot

const (
	// Cron job scheduling
	CronDailyJobCount = 1 // Run job once per day
	CronHourMidnight  = 0 // Hour: midnight (00:00:00)
	CronMinuteZero    = 0 // Minute: 0
	CronSecondZero    = 0 // Second: 0

	// Goroutine synchronization
	AddOneGoroutine = 1 // Increment wait group by 1 for single goroutine

	// Redis connection pool
	RedisPoolSize = 10 // Number of connections in Redis connection pool
)
