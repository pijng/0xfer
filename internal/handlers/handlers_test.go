package handlers

import (
	"testing"
)

func TestStringsHasPrefix(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{
			name:   "simple prefix match",
			s:      "hello world",
			prefix: "hello",
			want:   true,
		},
		{
			name:   "prefix not matching",
			s:      "hello world",
			prefix: "world",
			want:   false,
		},
		{
			name:   "empty string with empty prefix",
			s:      "",
			prefix: "",
			want:   true,
		},
		{
			name:   "empty prefix matches any string",
			s:      "hello",
			prefix: "",
			want:   true,
		},
		{
			name:   "longer prefix than string",
			s:      "hi",
			prefix: "hello",
			want:   false,
		},
		{
			name:   "exact match",
			s:      "test",
			prefix: "test",
			want:   true,
		},
		{
			name:   "case sensitive",
			s:      "Hello",
			prefix: "hello",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringsHasPrefix(tt.s, tt.prefix)
			if got != tt.want {
				t.Errorf("stringsHasPrefix(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500B",
		},
		{
			name:  "exactly 1KB",
			bytes: 1024,
			want:  "1.0KB",
		},
		{
			name:  "KB range",
			bytes: 1536,
			want:  "1.5KB",
		},
		{
			name:  "exactly 1MB",
			bytes: 1024 * 1024,
			want:  "1.0MB",
		},
		{
			name:  "MB range",
			bytes: 15 * 1024 * 1024,
			want:  "15.0MB",
		},
		{
			name:  "GB range",
			bytes: 2 * 1024 * 1024 * 1024,
			want:  "2.0GB",
		},
		{
			name:  "zero bytes",
			bytes: 0,
			want:  "0B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestEscapeQuotes(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "no quotes",
			s:    "hello world",
			want: "hello world",
		},
		{
			name: "single quote",
			s:    `hello"world`,
			want: `hello\"world`,
		},
		{
			name: "multiple quotes",
			s:    `"hello"`,
			want: `\"hello\"`,
		},
		{
			name: "empty string",
			s:    "",
			want: "",
		},
		{
			name: "only quote",
			s:    `"`,
			want: `\"`,
		},
		{
			name: "quotes at start and end",
			s:    `"file.txt"`,
			want: `\"file.txt\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeQuotes(tt.s)
			if got != tt.want {
				t.Errorf("escapeQuotes(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}