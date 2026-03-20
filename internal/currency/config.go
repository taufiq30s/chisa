package currency

const (
	// Currency Code Validation
	// CurrencyCodeLength is the standard length for ISO 4217 currency codes (e.g., USD, EUR, JPY)
	CurrencyCodeLength = 3

	// Amount Validation
	// MinimumAmount is the minimum allowed amount for currency conversion
	// Zero or negative amounts are not allowed
	MinimumAmount = 0

	// Rate Display Configuration
	// RateDisplayThreshold determines when to show more decimal places
	// If rate is less than 1, we show 5 decimal places for precision
	RateDisplayThreshold = 1

	// Time Format Constants
	// DateFormat is the format string for displaying dates in currency conversion results
	// Format: "02 January 2006 15:04:05 MST" (e.g., "19 March 2026 15:04:05 UTC")
	DateFormat = "02 January 2006 15:04:05 MST"

	// Handler Option Indices
	// These constants represent the indices for accessing command options in handlers
	FirstOptionIndex  = 0 // Primary command option
	SecondOptionIndex = 1 // Amount or secondary parameter
	ThirdOptionIndex  = 2 // Base currency
	FourthOptionIndex = 3 // Destination currency
)
