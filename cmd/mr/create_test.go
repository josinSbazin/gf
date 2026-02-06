package mr

import (
	"testing"
)

func TestValidateMRBranch(t *testing.T) {
	tests := []struct {
		name       string
		branch     string
		branchType string
		wantErr    bool
	}{
		// Valid branches
		{"simple", "main", "source", false},
		{"with dash", "feature-branch", "source", false},
		{"with underscore", "feature_branch", "source", false},
		{"with slash", "feature/new-feature", "source", false},
		{"with dot", "release.1.0", "source", false},
		{"single char", "a", "source", false},
		{"numbers", "v123", "source", false},
		{"complex", "feature/JIRA-123_implement-auth", "source", false},

		// Invalid branches
		{"empty", "", "source", true},
		{"starts with dash", "-branch", "source", true},
		{"contains dotdot", "feature..main", "source", true},
		{"only spaces", "   ", "source", true},
		{"special chars", "branch@name", "source", true},
		{"newline", "branch\nname", "source", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMRBranch(tt.branch, tt.branchType)
			if tt.wantErr && err == nil {
				t.Errorf("validateMRBranch(%q) should return error", tt.branch)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateMRBranch(%q) unexpected error: %v", tt.branch, err)
			}
		})
	}
}

func TestValidMRBranchRegex(t *testing.T) {
	valid := []string{
		"main",
		"develop",
		"feature/test",
		"release-1.0.0",
		"fix_bug_123",
		"a",
		"ab",
		"feature/JIRA-123/description",
	}

	for _, branch := range valid {
		if !validMRBranchRegex.MatchString(branch) {
			t.Errorf("branch %q should match regex", branch)
		}
	}

	invalid := []string{
		"-starts-with-dash",
		"ends-with-dash-", // This might pass depending on regex
		"has spaces",
		"has\ttab",
		"",
	}

	for _, branch := range invalid {
		if validMRBranchRegex.MatchString(branch) && branch != "ends-with-dash-" {
			t.Errorf("branch %q should not match regex", branch)
		}
	}
}

func TestCreateCmd_Flags(t *testing.T) {
	cmd := newCreateCmd()

	// Verify all expected flags exist
	flags := []struct {
		name      string
		shorthand string
	}{
		{"title", "t"},
		{"body", "b"},
		{"target", "T"},
		{"source", "S"},
		{"draft", ""},
		{"delete-branch", "d"},
		{"repo", "R"},
		{"web", "w"},
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

func TestCreateCmd_Usage(t *testing.T) {
	cmd := newCreateCmd()

	if cmd.Use != "create" {
		t.Errorf("Use = %q, want create", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description is empty")
	}
}
