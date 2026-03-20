package utils

const (
	// File permissions (octal values)
	FilePermissionReadWrite      = 0644 // -rw-r--r-- (owner: read/write, group: read, others: read)
	FilePermissionReadWriteGroup = 0664 // -rw-rw-r-- (owner: read/write, group: read/write, others: read)

	// Number formatting constants
	DecimalPlaces          = 2 // Number of decimal places for float formatting
	DecimalSeparatorOffset = 3 // Position offset for decimal separator in formatted number
	DigitsPerCommaGroup    = 3 // Number of digits between comma separators (e.g., 1,000,000)
	CommaGroupCalculator   = 1 // Used in calculation: (numOfDigits - 1) / 3

	// String indexing
	FirstCharIndex = 0 // Index of first character in string
	StringIndexOne = 1 // Offset for skipping first character (e.g., negative sign)

	// Loop boundaries
	LoopZeroBoundary = 0 // Boundary check for loop termination
)
