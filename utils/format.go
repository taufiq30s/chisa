package utils

import (
	"fmt"
)

func FormatDecimalNumberWithGrouping(num float64) string {
	in := fmt.Sprintf("%.2f", num)
	numOfDigits := len(in) - DecimalSeparatorOffset
	if num < LoopZeroBoundary {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - CommaGroupCalculator) / DigitsPerCommaGroup

	out := make([]byte, len(in)+numOfCommas)
	if num < LoopZeroBoundary {
		in, out[FirstCharIndex] = in[StringIndexOne:], '-'
	}

	isBehindDecimal := true
	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if out[j] == '.' {
			isBehindDecimal = false
			continue
		}
		if !isBehindDecimal {
			if i == LoopZeroBoundary {
				return string(out)
			}
			if k++; k == DigitsPerCommaGroup {
				j, k = j-CommaGroupCalculator, LoopZeroBoundary
				out[j] = ','
			}
		}
	}
}
