package cmd

import (
	"testing"
)

func TestBrowseOptions_Flags(t *testing.T) {
	cmd := newBrowseCmd()

	// Verify all expected flags exist
	flags := []struct {
		name      string
		shorthand string
	}{
		{"repo", "R"},
		{"branch", "b"},
		{"settings", "s"},
		{"issues", ""},
		{"mrs", ""},
		{"mr", "m"},
		{"pipeline", "p"},
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

func TestBrowseCmd_Args(t *testing.T) {
	cmd := newBrowseCmd()

	// Test that command accepts 0 or 1 argument
	if err := cmd.Args(cmd, []string{}); err != nil {
		t.Errorf("should accept 0 args: %v", err)
	}

	if err := cmd.Args(cmd, []string{"42"}); err != nil {
		t.Errorf("should accept 1 arg: %v", err)
	}

	if err := cmd.Args(cmd, []string{"42", "extra"}); err == nil {
		t.Error("should reject 2 args")
	}
}

func TestBrowseCmd_Usage(t *testing.T) {
	cmd := newBrowseCmd()

	if cmd.Use != "browse [<number>]" {
		t.Errorf("Use = %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cmd.Example == "" {
		t.Error("Example is empty")
	}
}
