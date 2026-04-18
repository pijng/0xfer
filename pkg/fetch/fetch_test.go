package fetch

import "testing"

func TestFilenameFromHeader(t *testing.T) {
	tests := []struct {
		name       string
		cd        string
		remotePath string
		want     string
	}{
		{
			name:       "quoted filename",
			cd:        `attachment; filename="test.zip"`,
			remotePath: "",
			want:     "test.zip",
		},
		{
			name:       "unquoted filename",
			cd:        "attachment; filename=test.zip",
			remotePath: "",
			want:     "test.zip",
		},
		{
			name:       "quoted with semicolon",
			cd:        `attachment; filename="test.zip"; size=123`,
			remotePath: "",
			want:     "test.zip",
		},
		{
			name:       "unquoted with semicolon",
			cd:        "attachment; filename=test.zip; size=123",
			remotePath: "",
			want:     "test.zip",
		},
		{
			name:       "from path",
			cd:        "",
			remotePath: "https://example.com/files/test.zip",
			want:     "test.zip",
		},
		{
			name:       "path with query",
			cd:        "",
			remotePath: "https://example.com/files/test.zip?token=abc",
			want:     "test.zip",
		},
		{
			name:       "empty cd uses path",
			cd:        "attachment",
			remotePath: "https://example.com/file.zip",
			want:     "file.zip",
		},
		{
			name:       "both present prefers cd",
			cd:        `attachment; filename="from-cd.zip"`,
			remotePath: "https://example.com/from-path.zip",
			want:     "from-cd.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilenameFromHeader(tt.cd, tt.remotePath)
			if got != tt.want {
				t.Errorf("FilenameFromHeader(%q, %q) = %q, want %q", tt.cd, tt.remotePath, got, tt.want)
			}
		})
	}
}