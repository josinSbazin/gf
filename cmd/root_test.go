package cmd

import (
	"testing"
)

func TestRootCmd_SubCommands(t *testing.T) {
	// Verify root command has all expected subcommands
	subCommands := []string{
		"api",
		"auth",
		"browse",
		"issue",
		"mr",
		"pipeline",
		"release",
		"repo",
		"version",
	}

	for _, name := range subCommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

func TestRootCmd_UseLine(t *testing.T) {
	if rootCmd.Use != "gf" {
		t.Errorf("Use = %q, want gf", rootCmd.Use)
	}
}

func TestRootCmd_Version(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("Version is empty")
	}
}

func TestRootCmd_HasShortDescription(t *testing.T) {
	if rootCmd.Short == "" {
		t.Error("Short description is empty")
	}
}

func TestRootCmd_HasLongDescription(t *testing.T) {
	if rootCmd.Long == "" {
		t.Error("Long description is empty")
	}
}

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCmd()

	if cmd.Use != "version" {
		t.Errorf("Use = %q, want version", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	// Verify Run function exists (not RunE)
	if cmd.Run == nil {
		t.Error("Run function is nil")
	}
}
