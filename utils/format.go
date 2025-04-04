package utils

import (
	"fmt"
)

func FormatDecimalNumberWithGrouping(num float64) string {
	in := fmt.Sprintf("%.2f", num)
	numOfDigits := len(in) - 3
	if num < 0 {
		numOfDigits-- // First character is the - sign (not a digit)
	}
	numOfCommas := (numOfDigits - 1) / 3

	out := make([]byte, len(in)+numOfCommas)
	if num < 0 {
		in, out[0] = in[1:], '-'
	}

	isBehindDecimal := true
	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if out[j] == '.' {
			isBehindDecimal = false
			continue
		}
		if !isBehindDecimal {
			if i == 0 {
				return string(out)
			}
			if k++; k == 3 {
				j, k = j-1, 0
				out[j] = ','
			}
		}
	}
}
