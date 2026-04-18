package config

import (
	"testing"
	"time"
)

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		def      string
		setup    func()
		teardown func()
		want     string
	}{
		{
			name:     "returns default when env not set",
			key:      "TEST_NONEXISTENT_VAR_12345",
			def:      "default_value",
			setup:    nil,
			teardown: nil,
			want:     "default_value",
		},
		{
			name: "returns env value when set",
			key:  "TEST_PARSE_SIZE_12345",
			def:  "default",
			setup: func() {
				t.Setenv("TEST_PARSE_SIZE_12345", "env_value")
			},
			teardown: nil,
			want:     "env_value",
		},
		{
			name: "empty env returns default",
			key:  "TEST_EMPTY_VAR_12345",
			def:  "default",
			setup: func() {
				t.Setenv("TEST_EMPTY_VAR_12345", "")
			},
			teardown: nil,
			want:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.teardown != nil {
				defer tt.teardown()
			}
			got := getEnvOrDefault(tt.key, tt.def)
			if got != tt.want {
				t.Errorf("getEnvOrDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name:  "empty returns default (100MB)",
			input: "",
			want:  100 << 20,
		},
		{
			name:  "parses bytes",
			input: "512",
			want:  512,
		},
		{
			name:  "parses KB",
			input: "100KB",
			want:  100 * 1024,
		},
		{
			name:  "parses MB",
			input: "50MB",
			want:  50 * 1024 * 1024,
		},
		{
			name:  "parses GB",
			input: "2GB",
			want:  2 * 1024 * 1024 * 1024,
		},
		{
			name:  "parses lowercase kb",
			input: "100kb",
			want:  100 * 1024,
		},
		{
			name:  "parses lowercase mb",
			input: "100mb",
			want:  100 * 1024 * 1024,
		},
		{
			name:  "parses lowercase gb",
			input: "1gb",
			want:  1 * 1024 * 1024 * 1024,
		},
		{
			name:  "invalid returns 0",
			input: "invalid",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSize(tt.input)
			if got != tt.want {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDur(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{
			name:  "empty returns default (168h)",
			input: "",
			want:  168 * time.Hour,
		},
		{
			name:  "parses hours",
			input: "24h",
			want:  24 * time.Hour,
		},
		{
			name:  "parses minutes",
			input: "30m",
			want:  30 * time.Minute,
		},
		{
			name:  "parses days",
			input: "7d",
			want:  7 * 24 * time.Hour,
		},
		{
			name:  "invalid returns default",
			input: "invalid",
			want:  168 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDur(tt.input)
			if got != tt.want {
				t.Errorf("parseDur(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}