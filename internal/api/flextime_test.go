package api

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFlexTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(FlexTime) bool
	}{
		{
			name:    "RFC3339 with timezone",
			input:   `"2024-02-05T18:17:51.081636Z"`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.Year() == 2024 && ft.Month() == 2 && ft.Day() == 5
			},
		},
		{
			name:    "RFC3339 with offset",
			input:   `"2024-02-05T18:17:51+03:00"`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.Year() == 2024 && ft.Month() == 2
			},
		},
		{
			name:    "without timezone microseconds",
			input:   `"2024-02-05T18:17:51.081636"`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.Year() == 2024 && ft.Hour() == 18
			},
		},
		{
			name:    "without timezone no microseconds",
			input:   `"2024-02-05T18:17:51"`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.Minute() == 17 && ft.Second() == 51
			},
		},
		{
			name:    "null value",
			input:   `"null"`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.IsZero()
			},
		},
		{
			name:    "empty string",
			input:   `""`,
			wantErr: false,
			check: func(ft FlexTime) bool {
				return ft.IsZero()
			},
		},
		{
			name:    "invalid format",
			input:   `"not-a-date"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexTime
			err := json.Unmarshal([]byte(tt.input), &ft)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.check != nil && !tt.check(ft) {
				t.Errorf("check failed for time: %v", ft.Time)
			}
		})
	}
}

func TestFlexTime_InStruct(t *testing.T) {
	type TestStruct struct {
		CreatedAt  FlexTime  `json:"createdAt"`
		FinishedAt *FlexTime `json:"finishedAt,omitempty"`
	}

	input := `{
		"createdAt": "2024-02-05T18:17:51.081636",
		"finishedAt": "2024-02-05T19:00:00Z"
	}`

	var ts TestStruct
	if err := json.Unmarshal([]byte(input), &ts); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if ts.CreatedAt.Year() != 2024 {
		t.Errorf("CreatedAt year = %d, want 2024", ts.CreatedAt.Year())
	}

	if ts.FinishedAt == nil {
		t.Fatal("FinishedAt is nil")
	}

	if ts.FinishedAt.Hour() != 19 {
		t.Errorf("FinishedAt hour = %d, want 19", ts.FinishedAt.Hour())
	}
}

func TestFlexTime_NilPointer(t *testing.T) {
	type TestStruct struct {
		FinishedAt *FlexTime `json:"finishedAt,omitempty"`
	}

	input := `{"finishedAt": null}`

	var ts TestStruct
	if err := json.Unmarshal([]byte(input), &ts); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// With null, the pointer should remain nil
	if ts.FinishedAt != nil && !ts.FinishedAt.IsZero() {
		t.Errorf("FinishedAt should be nil or zero, got %v", ts.FinishedAt)
	}
}

func TestFlexTime_AccessTimeField(t *testing.T) {
	var ft FlexTime
	json.Unmarshal([]byte(`"2024-02-05T12:30:45Z"`), &ft)

	// Test that we can access .Time field directly
	stdTime := ft.Time
	if stdTime.Hour() != 12 {
		t.Errorf("Time.Hour() = %d, want 12", stdTime.Hour())
	}

	// Test that time.Time methods are accessible via embedding
	if ft.Hour() != 12 {
		t.Errorf("Hour() = %d, want 12", ft.Hour())
	}

	// Test formatting
	formatted := ft.Format(time.RFC3339)
	if formatted != "2024-02-05T12:30:45Z" {
		t.Errorf("Format() = %s, want 2024-02-05T12:30:45Z", formatted)
	}
}
