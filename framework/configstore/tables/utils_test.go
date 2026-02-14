package tables

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     time.Duration
		wantErr  bool
	}{
		// Standard Go durations
		{
			name:     "30 seconds",
			duration: "30s",
			want:     30 * time.Second,
			wantErr:  false,
		},
		{
			name:     "5 minutes",
			duration: "5m",
			want:     5 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "1 hour",
			duration: "1h",
			want:     1 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "90 minutes",
			duration: "90m",
			want:     90 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "4 hours",
			duration: "4h",
			want:     4 * time.Hour,
			wantErr:  false,
		},

		// Days
		{
			name:     "1 day",
			duration: "1d",
			want:     24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "7 days",
			duration: "7d",
			want:     7 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "30 days",
			duration: "30d",
			want:     30 * 24 * time.Hour,
			wantErr:  false,
		},

		// Weeks
		{
			name:     "1 week",
			duration: "1w",
			want:     7 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "2 weeks",
			duration: "2w",
			want:     2 * 7 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "4 weeks",
			duration: "4w",
			want:     4 * 7 * 24 * time.Hour,
			wantErr:  false,
		},

		// Months (approximated as 30 days)
		{
			name:     "1 month",
			duration: "1M",
			want:     30 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "3 months",
			duration: "3M",
			want:     3 * 30 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "12 months",
			duration: "12M",
			want:     12 * 30 * 24 * time.Hour,
			wantErr:  false,
		},

		// Years (approximated as 365 days)
		{
			name:     "1 year lowercase",
			duration: "1y",
			want:     365 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "1 year uppercase",
			duration: "1Y",
			want:     365 * 24 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "2 years",
			duration: "2y",
			want:     2 * 365 * 24 * time.Hour,
			wantErr:  false,
		},

		// Complex flexible durations
		{
			name:     "90 minutes - flexible format",
			duration: "90m",
			want:     90 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "4 hours - flexible format",
			duration: "4h",
			want:     4 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "45 days - flexible format",
			duration: "45d",
			want:     45 * 24 * time.Hour,
			wantErr:  false,
		},

		// Error cases
		{
			name:     "empty duration",
			duration: "",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "invalid format - no unit",
			duration: "30",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "invalid format - no number",
			duration: "h",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "zero value",
			duration: "0h",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "negative value",
			duration: "-1h",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "too short",
			duration: "h",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.duration)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.duration, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestParseDuration_BackwardCompatibility(t *testing.T) {
	// Test that old duration formats still work
	tests := []struct {
		name     string
		duration string
		want     time.Duration
	}{
		{
			name:     "legacy 1m",
			duration: "1m",
			want:     1 * time.Minute,
		},
		{
			name:     "legacy 5m",
			duration: "5m",
			want:     5 * time.Minute,
		},
		{
			name:     "legacy 1h",
			duration: "1h",
			want:     1 * time.Hour,
		},
		{
			name:     "legacy 6h",
			duration: "6h",
			want:     6 * time.Hour,
		},
		{
			name:     "legacy 1d",
			duration: "1d",
			want:     24 * time.Hour,
		},
		{
			name:     "legacy 1w",
			duration: "1w",
			want:     7 * 24 * time.Hour,
		},
		{
			name:     "legacy 1M",
			duration: "1M",
			want:     30 * 24 * time.Hour,
		},
		{
			name:     "legacy 1Y",
			duration: "1Y",
			want:     365 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.duration)
			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error = %v", tt.duration, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

func TestParseDuration_FlexibleValues(t *testing.T) {
	// Test that flexible values work correctly
	tests := []struct {
		duration string
		want     time.Duration
	}{
		// Non-standard but valid values
		{"15m", 15 * time.Minute},
		{"90m", 90 * time.Minute},
		{"4h", 4 * time.Hour},
		{"8h", 8 * time.Hour},
		{"12h", 12 * time.Hour},
		{"36h", 36 * time.Hour},
		{"45d", 45 * 24 * time.Hour},
		{"60d", 60 * 24 * time.Hour},
		{"90d", 90 * 24 * time.Hour},
		{"3w", 3 * 7 * 24 * time.Hour},
		{"6M", 6 * 30 * 24 * time.Hour},
		{"24M", 24 * 30 * 24 * time.Hour},
		{"5y", 5 * 365 * 24 * time.Hour},
		{"10y", 10 * 365 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.duration, func(t *testing.T) {
			got, err := ParseDuration(tt.duration)
			if err != nil {
				t.Errorf("ParseDuration(%q) unexpected error = %v", tt.duration, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}
