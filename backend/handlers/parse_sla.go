package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

func parseSLA(value string) (any, any) {
	match := slaPattern.FindStringSubmatch(value)
	if len(match) < 3 {
		return nil, nil
	}
	number, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
	if err != nil {
		return nil, nil
	}
	unit := strings.ToLower(match[2])
	switch unit {
	case "minute", "minutes":
		unit = "menit"
	case "hour", "hours":
		unit = "jam"
	case "day", "days":
		unit = "hari"
	case "week", "weeks":
		unit = "minggu"
	}
	return number, unit
}

func formatSLA(value sql.NullFloat64, unit sql.NullString) string {
	if !value.Valid || !unit.Valid || unit.String == "" {
		return ""
	}
	if value.Float64 == float64(int64(value.Float64)) {
		return fmt.Sprintf("%d %s", int64(value.Float64), unit.String)
	}
	return fmt.Sprintf("%.1f %s", value.Float64, unit.String)
}
