package main

import "time"

const (
	// Goroutine synchronization
	NumInitGoroutines = 4 // Number of goroutines to wait for during initialization (music client, currency client, quiz service, handlers)

	// Signal handling
	SignalChannelBuffer = 1 // Buffer size for OS signal channel to prevent blocking

	// Date/Time formatting
	// Reference time: Mon Jan 2 15:04:05 MST 2006
	DateTimeFormat = "Mon Jan 2 2006 15:04:05 GMT+0000" // Format for displaying bot uptime

	// Shutdown
	ShutdownTimeout = 10 * time.Second // Maximum time to wait for graceful shutdown
)
