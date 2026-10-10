package server

import (
	"strings"
	"time"
)

// Some players append a twelve-hour clock marker to the timeshift path.
// Providers expect the same local time expressed with a twenty-four-hour clock.
func normalizeTimeshiftStart(start string) string {
	value := strings.ToUpper(strings.TrimSpace(start))
	marker := ""
	if strings.HasSuffix(value, "AM") || strings.HasSuffix(value, "PM") {
		marker = " " + value[len(value)-2:]
		value = strings.TrimSpace(value[:len(value)-2])
	}
	for _, dateSeparator := range []string{":", " ", "T"} {
		for _, clock := range []string{"15-04", "15:04", "15:04:05", "15-04-05"} {
			layout := "2006-01-02" + dateSeparator + clock
			if marker != "" {
				layout = strings.Replace(layout, "15", "3", 1) + " PM"
			}
			if parsed, err := time.Parse(layout, value+marker); err == nil {
				if strings.Contains(clock, "05") {
					return parsed.Format("2006-01-02:15:04:05")
				}
				return parsed.Format("2006-01-02:15-04")
			}
		}
	}
	return start
}
