package utils

import (
	"fmt"
	"regexp"
	"strconv"
)

var iso8601DurationPattern = regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?$`)

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

func FormatISO8601ToSeconds(iso8601 string) (int, error) {
	matches := iso8601DurationPattern.FindStringSubmatch(iso8601)
	if matches == nil {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", iso8601)
	}

	var hours, minutes, seconds int
	var err error

	if matches[1] != "" {
		hours, err = strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("invalid hours in ISO 8601 duration: %s", iso8601)
		}
	}
	if matches[2] != "" {
		minutes, err = strconv.Atoi(matches[2])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in ISO 8601 duration: %s", iso8601)
		}
	}
	if matches[3] != "" {
		seconds, err = strconv.Atoi(matches[3])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds in ISO 8601 duration: %s", iso8601)
		}
	}

	if hours == 0 && minutes == 0 && seconds == 0 {
		return 0, fmt.Errorf("invalid ISO 8601 duration format: %s", iso8601)
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	return totalSeconds, nil
}
