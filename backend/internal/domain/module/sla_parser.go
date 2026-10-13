package module

import (
	"regexp"
	"strconv"
	"strings"
)

var slaPattern = regexp.MustCompile(`(?i)(\d+(?:[.,]\d+)?)\s*(menit|minute|minutes|jam|hour|hours|hari|day|days|minggu|week|weeks)`)

// ParseSLA parses a SLA text string into an integer duration and standardized unit.
// Supported units are menit, jam, hari, minggu.
func ParseSLA(slaText string) (int, string) {
	val, unit := ParseSLAFloat(slaText)
	return int(val), unit
}

// ParseSLAFloat parses a SLA text string into a float64 duration and standardized unit.
func ParseSLAFloat(slaText string) (float64, string) {
	match := slaPattern.FindStringSubmatch(slaText)
	if len(match) < 3 {
		return 0, ""
	}
	number, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return 0, ""
	}
	unit := normalizeSLAUnit(match[2])
	return number, unit
}

// ParseSLAAny returns any, any representation for database upserts (nil when empty).
func ParseSLAAny(slaText string) (any, any) {
	match := slaPattern.FindStringSubmatch(slaText)
	if len(match) < 3 {
		return nil, nil
	}
	number, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return nil, nil
	}
	unit := normalizeSLAUnit(match[2])
	return number, unit
}

func normalizeSLAUnit(rawUnit string) string {
	switch strings.ToLower(rawUnit) {
	case "minute", "minutes":
		return "menit"
	case "hour", "hours":
		return "jam"
	case "day", "days":
		return "hari"
	case "week", "weeks":
		return "minggu"
	default:
		return strings.ToLower(rawUnit)
	}
}
