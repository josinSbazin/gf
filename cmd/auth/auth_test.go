package auth

import (
	"testing"
)

func TestAuthCmd_SubCommands(t *testing.T) {
	cmd := NewCmdAuth()

	// Verify auth command has all expected subcommands
	subCommands := []string{
		"login",
		"logout",
		"status",
	}

	for _, name := range subCommands {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestAuthCmd_Usage(t *testing.T) {
	cmd := NewCmdAuth()

	if cmd.Use != "auth" {
		t.Errorf("Use = %q, want auth", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestLoginCmd_Flags(t *testing.T) {
	cmd := newLoginCmd()

	flags := []struct {
		name      string
		shorthand string
	}{
		{"hostname", "H"},
		{"token", "t"},
		{"stdin", ""},
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
