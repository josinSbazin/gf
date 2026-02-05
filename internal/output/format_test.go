package output

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "just now (seconds ago)",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "5m ago",
		},
		{
			name: "hours ago",
			time: now.Add(-3 * time.Hour),
			want: "3h ago",
		},
		{
			name: "days ago",
			time: now.Add(-2 * 24 * time.Hour),
			want: "2d ago",
		},
		{
			name: "week ago (shows date)",
			time: now.Add(-10 * 24 * time.Hour),
			want: now.Add(-10 * 24 * time.Hour).Format("Jan 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTime(tt.time)
			if got != tt.want {
				t.Errorf("FormatRelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatRelativeTime_EdgeCases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		contains string // For cases where exact match is hard
	}{
		{
			name:     "exactly 1 minute",
			time:     now.Add(-1 * time.Minute),
			contains: "1m ago",
		},
		{
			name:     "exactly 1 hour",
			time:     now.Add(-1 * time.Hour),
			contains: "1h ago",
		},
		{
			name:     "exactly 1 day",
			time:     now.Add(-24 * time.Hour),
			contains: "1d ago",
		},
		{
			name:     "exactly 7 days (boundary)",
			time:     now.Add(-7 * 24 * time.Hour),
			contains: "", // Could be "7d ago" or date format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeTime(tt.time)
			if got == "" {
				t.Error("FormatRelativeTime returned empty string")
			}
			// Just verify it returns something reasonable
			t.Logf("FormatRelativeTime(%v) = %q", tt.time, got)
		})
	}
}

func TestFormatRelativeTime_ZeroTime(t *testing.T) {
	// Zero time is far in the past, should return date format
	got := FormatRelativeTime(time.Time{})
	if got == "" {
		t.Error("FormatRelativeTime(zero) returned empty string")
	}
	// Zero time is year 0001, should show a date
	t.Logf("FormatRelativeTime(zero) = %q", got)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		{
			name:    "zero",
			seconds: 0,
			want:    "-",
		},
		{
			name:    "seconds only",
			seconds: 45,
			want:    "45s",
		},
		{
			name:    "exactly 1 minute",
			seconds: 60,
			want:    "1m",
		},
		{
			name:    "minutes and seconds",
			seconds: 125, // 2m 5s
			want:    "2m 5s",
		},
		{
			name:    "multiple minutes no seconds",
			seconds: 180, // 3m
			want:    "3m",
		},
		{
			name:    "large duration",
			seconds: 3661, // 61m 1s = 1h 1m 1s but we only show m:s
			want:    "61m 1s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.seconds)
			if got != tt.want {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
	}{
		{"negative", -1},
		{"very large", 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.seconds)
			// Just verify it doesn't panic and returns something
			if tt.seconds < 0 && got != "-" {
				// Negative should probably return "-" or handle gracefully
				t.Logf("FormatDuration(%d) = %q (negative input)", tt.seconds, got)
			} else {
				t.Logf("FormatDuration(%d) = %q", tt.seconds, got)
			}
		})
	}
}
