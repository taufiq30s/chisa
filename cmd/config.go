package main

const (
	// Goroutine synchronization
	NumInitGoroutines = 3 // Number of goroutines to wait for during initialization (music client, currency client, handlers)

	// Signal handling
	SignalChannelBuffer = 1 // Buffer size for OS signal channel to prevent blocking

	// Date/Time formatting
	// Reference time: Mon Jan 2 15:04:05 MST 2006
	DateTimeFormat = "Mon Jan 2 2006 15:04:05 GMT+0000" // Format for displaying bot uptime
)
