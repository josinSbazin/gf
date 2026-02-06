package cmd

import (
	"encoding/json"
	"testing"
)

func TestSimpleJQ(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		filter  string
		want    string
		wantErr bool
	}{
		{
			name:   "identity",
			input:  `{"foo": "bar"}`,
			filter: ".",
			want:   `{"foo": "bar"}`,
		},
		{
			name:   "simple field",
			input:  `{"name": "test", "value": 123}`,
			filter: ".name",
			want:   `"test"`,
		},
		{
			name:   "nested field",
			input:  `{"user": {"name": "alice"}}`,
			filter: ".user.name",
			want:   `"alice"`,
		},
		{
			name:   "array index",
			input:  `[1, 2, 3]`,
			filter: ".[0]",
			want:   `1`,
		},
		{
			name:   "field with array index",
			input:  `{"items": ["a", "b", "c"]}`,
			filter: ".items[1]",
			want:   `"b"`,
		},
		{
			name:   "number value",
			input:  `{"count": 42}`,
			filter: ".count",
			want:   `42`,
		},
		{
			name:   "boolean value",
			input:  `{"active": true}`,
			filter: ".active",
			want:   `true`,
		},
		{
			name:   "null value",
			input:  `{"data": null}`,
			filter: ".data",
			want:   `null`,
		},
		{
			name:    "non-existent field",
			input:   `{"foo": "bar"}`,
			filter:  ".baz",
			want:    `null`,
			wantErr: false,
		},
		{
			name:    "array index out of bounds",
			input:   `[1, 2]`,
			filter:  ".[5]",
			wantErr: true,
		},
		{
			name:    "invalid array index",
			input:   `[1, 2]`,
			filter:  ".[abc]",
			wantErr: true,
		},
		{
			name:    "access field on non-object",
			input:   `"string"`,
			filter:  ".field",
			wantErr: true,
		},
		{
			name:    "index non-array",
			input:   `{"foo": "bar"}`,
			filter:  ".[0]",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := simpleJQ(json.RawMessage(tt.input), tt.filter)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Normalize JSON for comparison
			var gotVal, wantVal any
			if err := json.Unmarshal(got, &gotVal); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.want), &wantVal); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			gotBytes, _ := json.Marshal(gotVal)
			wantBytes, _ := json.Marshal(wantVal)

			if string(gotBytes) != string(wantBytes) {
				t.Errorf("simpleJQ() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestAPICmd_Flags(t *testing.T) {
	cmd := newAPICmd()

	// Verify all expected flags exist
	flags := []string{"method", "hostname", "header", "field", "raw-field", "input", "silent", "jq"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag --%s not found", name)
		}
	}
}

func TestAPICmd_Args(t *testing.T) {
	cmd := newAPICmd()

	// Requires exactly 1 argument
	if err := cmd.Args(cmd, []string{}); err == nil {
		t.Error("should reject 0 args")
	}

	if err := cmd.Args(cmd, []string{"/test"}); err != nil {
		t.Errorf("should accept 1 arg: %v", err)
	}

	if err := cmd.Args(cmd, []string{"/test", "extra"}); err == nil {
		t.Error("should reject 2 args")
	}
}

func TestValidHTTPMethods(t *testing.T) {
	valid := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for _, m := range valid {
		if !validHTTPMethods[m] {
			t.Errorf("method %s should be valid", m)
		}
	}

	invalid := []string{"INVALID", "get", "TRACE", "CONNECT"}
	for _, m := range invalid {
		if validHTTPMethods[m] {
			t.Errorf("method %s should be invalid", m)
		}
	}
}
