package tables

import (
	"fmt"
	"strconv"
	"time"
)

// ParseDuration parses a flexible duration string into a time.Duration.
//
// Supported formats:
//   - Standard Go durations: "30s", "5m", "1h" (via time.ParseDuration)
//   - Days: "1d", "7d", "30d" (converted to hours * 24)
//   - Weeks: "1w", "2w", "4w" (converted to hours * 24 * 7)
//   - Months: "1M", "3M", "12M" (approximated as 30 days)
//   - Years: "1y", "2Y" (either case, approximated as 365 days)
//
// Examples:
//   - "90m" -> 90 minutes
//   - "4h" -> 4 hours
//   - "7d" -> 7 days (168 hours)
//   - "2w" -> 2 weeks (336 hours)
//   - "1M" -> 1 month (720 hours, approximated)
//   - "1y" or "1Y" -> 1 year (8760 hours, approximated)
//
// This function is backward compatible with existing duration strings
// and supports the new flexible format allowing arbitrary numbers.
func ParseDuration(duration string) (time.Duration, error) {
	if duration == "" {
		return 0, fmt.Errorf("duration is empty")
	}

	// Get the numeric part and the unit
	if len(duration) < 2 {
		return 0, fmt.Errorf("duration too short: %s", duration)
	}

	unit := duration[len(duration)-1:]
	numericPart := duration[:len(duration)-1]

	// Parse the numeric value
	value, err := strconv.ParseFloat(numericPart, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric value in duration '%s': %w", duration, err)
	}

	if value <= 0 {
		return 0, fmt.Errorf("duration value must be positive: %s", duration)
	}

	// Handle special cases for days, weeks, months, years
	switch unit {
	case "d":
		// Days: convert to hours then multiply by 24
		return time.Duration(value * float64(24*time.Hour)), nil
	case "w":
		// Weeks: convert to hours then multiply by 24 * 7
		return time.Duration(value * float64(24*7*time.Hour)), nil
	case "M":
		// Months: approximate as 30 days
		return time.Duration(value * float64(24*30*time.Hour)), nil
	case "y", "Y":
		// Years: accept both lowercase and uppercase, approximate as 365 days
		return time.Duration(value * float64(24*365*time.Hour)), nil
	default:
		// For standard units (s, m, h, ms, etc.), use the standard parser
		// Reconstruct the duration string and let time.ParseDuration handle it
		return time.ParseDuration(duration)
	}
}
