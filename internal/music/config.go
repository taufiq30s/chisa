package music

import "time"

const (
	// Time Conversion Constants
	// SecondsPerMinute is used to convert seconds to minutes
	SecondsPerMinute = 60
	// MinutesPerHour is used to convert minutes to hours (not currently used but good to have)
	MinutesPerHour = 60

	// ResolveTimeout bounds a single Resolve call (yt-dlp search or a URL probe).
	ResolveTimeout = 20 * time.Second

	// Search Configuration
	// MaxSearchResults is the maximum number of search results to return from a search query
	MaxSearchResults = 20
	// SearchResultsPerPage is the number of search results to display per page
	SearchResultsPerPage = 5

	// Playlist Configuration
	// PlaylistResolveTimeout bounds a playlist extraction. It is far larger than
	// ResolveTimeout because yt-dlp walks every entry of the playlist. The
	// number of tracks one playlist may queue is capped by
	// providers.ytdlpPlaylistLimit, applied where the extraction happens.
	PlaylistResolveTimeout = 60 * time.Second

	// Queue Configuration
	// InitialQueueCapacity is the initial capacity when creating a new queue slice
	InitialQueueCapacity = 0

	// Pagination Configuration
	// InitialPage is the starting page number for search results
	InitialPage = 1

	// Queue List / Idle Configuration
	// QueueResultsPerPage is the number of queued tracks shown per /music list page
	QueueResultsPerPage = 10
	// AutoDisconnectDelay is how long the bot stays in the voice channel after
	// the queue empties. Change this value to tune the idle timeout; the
	// user-facing "No Activity" message is generated from it, so the text stays
	// correct.
	AutoDisconnectDelay = 10 * time.Minute
)
