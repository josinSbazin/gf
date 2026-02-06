package pipeline

import (
	"testing"
)

func TestWatchCmd_Flags(t *testing.T) {
	cmd := newWatchCmd()

	// Verify all expected flags exist
	flags := []struct {
		name      string
		shorthand string
	}{
		{"interval", "i"},
		{"exit-status", ""},
		{"repo", "R"},
	}

	for _, f := range flags {
		flag := cmd.Flags().Lookup(f.name)
		if flag == nil {
			t.Errorf("flag --%s not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag --%s shorthand = %q, want %q", f.name, flag.Shorthand, f.shorthand)
		}
	}
}

func TestWatchCmd_Args(t *testing.T) {
	cmd := newWatchCmd()

	// Requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("should reject 0 args")
	}

	if err := cmd.Args(cmd, []string{"42"}); err != nil {
		t.Errorf("should accept 1 arg: %v", err)
	}

	if err := cmd.Args(cmd, []string{"42", "extra"}); err == nil {
		t.Error("should reject 2 args")
	}
}

func TestIsFinished(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"success", true},
		{"passed", true},
		{"failed", true},
		{"canceled", true},
		{"running", false},
		{"pending", false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := isFinished(tt.status); got != tt.want {
				t.Errorf("isFinished(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIntervalConstants(t *testing.T) {
	if minInterval < 1 {
		t.Errorf("minInterval = %d, should be at least 1", minInterval)
	}

	if maxInterval <= minInterval {
		t.Errorf("maxInterval (%d) should be greater than minInterval (%d)", maxInterval, minInterval)
	}

	if maxInterval > 600 {
		t.Errorf("maxInterval = %d, seems too high", maxInterval)
	}
}
