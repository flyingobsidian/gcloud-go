package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// gcloudDurationUnits maps shorthand duration suffixes to their length.
var gcloudDurationUnits = map[byte]time.Duration{
	'd': 24 * time.Hour,
	'h': time.Hour,
	'm': time.Minute,
	's': time.Second,
}

// parseGcloudDuration parses a duration as accepted by Python gcloud's
// arg_parsers.Duration (see `gcloud topic datetimes`):
//   - a bare number, in seconds: "90", "1.5"
//   - shorthand with d/h/m/s units: "7d", "1h30m", "2.5h"
//   - ISO 8601 with days, hours, minutes, seconds: "P1D", "PT1H30M", "P1DT2H"
//
// Units are case-insensitive. Negative durations are rejected.
func parseGcloudDuration(s string) (time.Duration, error) {
	str := strings.ToLower(strings.TrimSpace(s))
	if str == "" {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	if secs, err := strconv.ParseFloat(str, 64); err == nil {
		if secs < 0 {
			return 0, fmt.Errorf("invalid duration %q: must not be negative", s)
		}
		return time.Duration(secs * float64(time.Second)), nil
	}
	if strings.HasPrefix(str, "p") {
		datePart, timePart, _ := strings.Cut(str[1:], "t")
		if strings.ContainsAny(datePart, "hms") {
			return 0, fmt.Errorf("invalid duration %q: only days are supported before T", s)
		}
		if datePart == "" && timePart == "" {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		str = datePart + timePart
	}
	var total time.Duration
	for str != "" {
		i := strings.IndexAny(str, "dhms")
		if i <= 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		n, err := strconv.ParseFloat(str[:i], 64)
		if err != nil || n < 0 {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		total += time.Duration(n * float64(gcloudDurationUnits[str[i]]))
		str = str[i+1:]
	}
	return total, nil
}
