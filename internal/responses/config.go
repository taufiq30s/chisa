package responses

const (
	// Default values
	DefaultVersion = "0.0.1" // Default bot version when VERSION env var is not set

	// Hex color conversion constants
	HexBase        = 16  // Base for hexadecimal number system
	HexDigitOffset = 10  // Offset to convert hex letter 'a'-'f' to decimal (a=10, b=11, etc.)
	HexCharZero    = '0' // ASCII value reference for '0'
	HexCharNine    = '9' // ASCII value reference for '9'
	HexCharLowerA  = 'a' // ASCII value reference for lowercase 'a'
	HexCharLowerF  = 'f' // ASCII value reference for lowercase 'f'

	// String manipulation
	EmptyString    = "" // Empty string for replacements
	NoReplaceLimit = -1 // No limit when replacing strings

	// Initialization values
	InitialValue = 0 // Initial value for counters/accumulator
	InitialBase  = 1 // Initial base value for number conversion

	// Component validation
	MinComponentCount = 0 // Minimum count for checking if components exist
)
