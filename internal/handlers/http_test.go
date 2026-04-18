package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "health check",
			method:     http.MethodGet,
			path:       "/health",
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			NewHealthHandler().ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d; want %d", w.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("got body %q; want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestDownloadHandlerNotFound(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "missing id",
			method:     http.MethodGet,
			path:       "/d/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "no such file",
			method:     http.MethodGet,
			path:       "/d/nonexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			NewDownloadHandler(nil).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestDeleteHandlerNotFound(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "missing id and secret",
			method:     http.MethodDelete,
			path:       "/d/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing secret",
			method:     http.MethodDelete,
			path:       "/d/abc123",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			NewDeleteHandler(nil).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d; want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUploadHandlerParseRequest(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		body         string
		wantFilename string
		wantSize     int64
	}{
		{
			name:         "binary body",
			contentType:  "application/octet-stream",
			body:         "hello world",
			wantFilename: "file",
			wantSize:     11,
		},
		{
			name:         "custom content type",
			contentType:  "image/png",
			body:         "fake png",
			wantFilename: "file",
			wantSize:     8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			handler := &UploadHandler{maxSize: 100} //nolint:exhaustruct
			filename, contentType, size, _, err := handler.parseRequest(req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if filename != tt.wantFilename {
				t.Errorf("got filename %q; want %q", filename, tt.wantFilename)
			}

			if contentType != tt.contentType {
				t.Errorf("got contentType %q; want %q", contentType, tt.contentType)
			}

			if size != tt.wantSize {
				t.Errorf("got size %d; want %d", size, tt.wantSize)
			}
		})
	}
}



func BenchmarkFormatSize(b *testing.B) {
	sizes := []int64{100, 1024, 1024 * 1024, 15 * 1024 * 1024, 2 * 1024 * 1024 * 1024}

	for _, size := range sizes {
		b.Run(string(rune('0'+size%9)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = formatSize(size)
			}
		})
	}
}

func BenchmarkEscapeQuotes(b *testing.B) {
	inputs := []string{
		"simple.txt",
		`file with "quotes".txt`,
		strings.Repeat("a", 1000),
		strings.Repeat(`"`, 100),
	}

	for _, input := range inputs {
		b.Run(input[:min(20, len(input))], func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = escapeQuotes(input)
			}
		})
	}
}

func BenchmarkStringsHasPrefix(b *testing.B) {
	inputs := []struct {
		s      string
		prefix string
	}{
		{"hello world", "hello"},
		{"hello world", "world"},
		{strings.Repeat("a", 1000), "aaa"},
		{"", ""},
	}

	for _, input := range inputs {
		b.Run(input.prefix, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = stringsHasPrefix(input.s, input.prefix)
			}
		})
	}
}

func BenchmarkHealthHandler(b *testing.B) {
	handler := NewHealthHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkDownloadHandlerNotFound(b *testing.B) {
	handler := NewDownloadHandler(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/d/nonexistent", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkDeleteHandlerNotFound(b *testing.B) {
	handler := NewDeleteHandler(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodDelete, "/d/id/secret", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkUploadHandlerWrongMethod(b *testing.B) {
	handler := NewUploadHandler(nil, 100, "http://localhost:2052")

	methods := []string{http.MethodGet, http.MethodPost, http.MethodDelete}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, method := range methods {
			req := httptest.NewRequest(method, "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

var _ io.Reader