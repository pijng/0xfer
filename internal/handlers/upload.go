package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"0xfer/internal/services"
)

type UploadHandler struct {
	service *services.FileService
	maxSize int64
	baseURL string
}

func NewUploadHandler(service *services.FileService, maxSize int64, baseURL string) http.Handler {
	h := &UploadHandler{service: service, maxSize: maxSize, baseURL: baseURL}

	return http.HandlerFunc(h.serve)
}

func (h *UploadHandler) serve(w http.ResponseWriter, r *http.Request) {
	filename, contentType, size, body, err := h.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if size > h.maxSize {
		http.Error(w, fmt.Sprintf("file too large: max %d bytes", h.maxSize), http.StatusBadRequest)

		return
	}

	result, err := h.service.Upload(r.Context(), filename, contentType, size, body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)

		return
	}

	downloadURL := fmt.Sprintf("%s/d/%s", h.baseURL, result.ID)
	deleteURL := fmt.Sprintf("%s/d/%s/%s", h.baseURL, result.ID, result.Secret)

	downloadCmd := fmt.Sprintf("curl -JLOf %s", downloadURL)
	deleteCmd := fmt.Sprintf("curl -X DELETE %s", deleteURL)
	expires := result.Expires.Round(time.Second).Format(time.RFC3339)

	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "File: %s (%s)\nDownload: %s\nDelete:   %s\nExpires:  %s\n",
		result.Filename, formatSize(result.Size),
		downloadCmd, deleteCmd, expires)
}

func (h *UploadHandler) parseRequest(r *http.Request) (filename, contentType string, size int64, body io.Reader, err error) {
	ct := r.Header.Get("Content-Type")

	if stringsHasPrefix(ct, "multipart/form-data") {
		return h.parseMultipart(r)
	}

	if len(ct) > 0 && ct != "application/octet-stream" && ct != "multipart/form-data" {
		filename = "file"
		contentType = ct
		size = r.ContentLength
		body = r.Body

		return
	}

	filename = "file"
	contentType = "application/octet-stream"
	size = r.ContentLength
	body = r.Body

	return
}

func (h *UploadHandler) parseMultipart(r *http.Request) (filename, contentType string, size int64, body io.Reader, err error) {
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		return "", "", 0, nil, fmt.Errorf("parse multipart form: %w", err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("get file from form: %w", err)
	}
	defer func() { _ = file.Close() }()

	return header.Filename, header.Header.Get("Content-Type"), header.Size, file, nil
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%dB", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMB", float64(bytes)/(1024*1024))
	}

	return fmt.Sprintf("%.1fGB", float64(bytes)/(1024*1024*1024))
}
