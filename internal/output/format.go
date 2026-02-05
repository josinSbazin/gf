package output

import (
	"fmt"
	"time"
)

// FormatRelativeTime formats a time as a relative string (e.g., "5m ago", "2d ago")
func FormatRelativeTime(t time.Time) string {
	diff := time.Since(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}

// FormatDuration formats seconds as a human-readable duration
func FormatDuration(seconds int) string {
	if seconds == 0 {
		return "-"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	mins := seconds / 60
	secs := seconds % 60
	if secs == 0 {
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%dm %ds", mins, secs)
}
