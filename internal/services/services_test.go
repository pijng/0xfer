package services

import (
	"testing"
	"time"

	"github.com/oklog/ulid"
)

func TestToBase62(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		wantOK func(string) bool
	}{
		{
			name:   "empty slice",
			input:  []byte{},
			wantOK: func(s string) bool { return s == "" },
		},
		{
			name:   "single byte",
			input:  []byte{0},
			wantOK: func(s string) bool { return len(s) == 1 },
		},
		{
			name:   "all zeros",
			input:  []byte{0, 0, 0, 0},
			wantOK: func(s string) bool { return len(s) == 4 },
		},
		{
			name:   "max byte value",
			input:  []byte{255},
			wantOK: func(s string) bool { return len(s) == 1 },
		},
		{
			name:   "8 bytes",
			input:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantOK: func(s string) bool { return len(s) == 8 },
		},
		{
			name:   "16 bytes",
			input:  make([]byte, 16),
			wantOK: func(s string) bool { return len(s) == 16 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toBase62(tt.input)
			if !tt.wantOK(got) {
				t.Errorf("toBase62(%v) = %q, want length %d", tt.input, got, len(tt.input))
			}
		})
	}
}

func TestBase62Chars(t *testing.T) {
	if len(base62Chars) != 62 {
		t.Errorf("base62Chars length = %d, want 62", len(base62Chars))
	}

	for i := 0; i < 26; i++ {
		if base62Chars[i] < 'a' || base62Chars[i] > 'z' {
			t.Errorf("base62Chars[%d] = %c, expected lowercase letter", i, base62Chars[i])
		}
	}

	for i := 26; i < 52; i++ {
		if base62Chars[i] < 'A' || base62Chars[i] > 'Z' {
			t.Errorf("base62Chars[%d] = %c, expected uppercase letter", i, base62Chars[i])
		}
	}

	for i := 52; i < 62; i++ {
		if base62Chars[i] < '0' || base62Chars[i] > '9' {
			t.Errorf("base62Chars[%d] = %c, expected digit", i, base62Chars[i])
		}
	}
}

func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateID()
	}
}

func BenchmarkGenerateSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateSecret()
	}
}

func BenchmarkToBase62(b *testing.B) {
	inputs := [][]byte{
		{1, 2, 3, 4, 5, 6, 7, 8},
		{255, 254, 253, 252, 251, 250, 249, 248},
		make([]byte, 16),
	}

	for _, input := range inputs {
		b.Run(string(input[:4]), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = toBase62(input)
			}
		})
	}
}

func BenchmarkULIDParse(b *testing.B) {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = id.String()[:10]
	}
}