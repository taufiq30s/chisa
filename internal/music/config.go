package music

import "time"

const (
	// Connection and Timeout Configuration
	// NodeConnectionTimeout is the maximum time to wait when connecting to a lavalink node
	NodeConnectionTimeout = 10 * time.Second

	// Time Conversion Constants
	// MillisecondsPerSecond is used to convert milliseconds to seconds
	MillisecondsPerSecond = 1000
	// SecondsPerMinute is used to convert seconds to minutes
	SecondsPerMinute = 60
	// MinutesPerHour is used to convert minutes to hours (not currently used but good to have)
	MinutesPerHour = 60

	// Search Configuration
	// MaxSearchResults is the maximum number of search results to return from a search query
	MaxSearchResults = 20
	// SearchResultsPerPage is the number of search results to display per page
	SearchResultsPerPage = 5

	// Queue Configuration
	// InitialQueueCapacity is the initial capacity when creating a new queue slice
	InitialQueueCapacity = 0
	// InitialLastPosition indicates that no track has been played yet
	InitialLastPosition = -1
	// FirstQueueIndex is the index of the first track in the queue
	FirstQueueIndex = 0
	// SecondQueueIndex is the index of the second track in the queue
	SecondQueueIndex = 1
	// QueueOffsetForLength represents the offset to calculate the actual queue length
	// (current track is not counted in queue length display)
	QueueOffsetForLength = 1

	// Pagination Configuration
	// InitialPage is the starting page number for search results
	InitialPage = 1

	// Wait/Sleep Durations
	// InteractionResponseDelay is the delay before sending an interaction response
	// This gives Discord time to process the deferred response
	InteractionResponseDelay = 500 * time.Millisecond

	// Search Cache Configuration
	// SearchCacheCleanupInterval is how often to check for expired search cache entries
	SearchCacheCleanupInterval = 1 * time.Minute
	// SearchCacheExpiryDuration is how long a search cache entry is valid
	SearchCacheExpiryDuration = 1 * time.Minute

	// WaitGroup Configuration
	// WaitGroupIncrement is the number to add to wait group when adding a goroutine
	WaitGroupIncrement = 1
)
