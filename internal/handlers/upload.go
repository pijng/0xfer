package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"0xfer/internal/services"
	"0xfer/pkg/fetch"
	"0xfer/pkg/netsec"
)

type UploadHandler struct {
	service    *services.FileService
	maxSize    int64
	baseURL    string
	maxTTL     time.Duration
	httpClient *http.Client
}

type UploadOptions struct {
	URL          string
	Expires      time.Duration
	MaxDownloads int
}

func NewUploadHandler(service *services.FileService, maxSize int64, baseURL string, maxTTL time.Duration) http.Handler {
	h := &UploadHandler{
		service: service,
		maxSize: maxSize,
		baseURL: baseURL,
		maxTTL:  maxTTL,
		httpClient: &http.Client{ //nolint:exhaustruct
			Timeout: 1 * time.Minute,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	return http.HandlerFunc(h.serve)
}

func (h *UploadHandler) serve(w http.ResponseWriter, r *http.Request) {
	filename, contentType, size, body, opts, err := h.parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	if opts.URL != "" {
		filename, contentType, size, body, err = h.fetchRemote(opts.URL)
		if err != nil {
			http.Error(w, fmt.Sprintf("fetchurl: %s", err.Error()), http.StatusBadRequest)

			return
		}
	}

	if size > h.maxSize {
		http.Error(w, fmt.Sprintf("file too large: max %d bytes", h.maxSize), http.StatusBadRequest)

		return
	}

	ttl := h.maxTTL
	if opts.Expires > 0 && opts.Expires < ttl {
		ttl = opts.Expires
	}

	result, err := h.service.Upload(r.Context(), filename, contentType, size, body, ttl, opts.MaxDownloads)
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

func (h *UploadHandler) parseRequest(r *http.Request) (filename, contentType string, size int64, body io.Reader, opts *UploadOptions, err error) {
	opts = &UploadOptions{} //nolint:exhaustruct

	if err := r.ParseForm(); err == nil {
		if remoteURL := r.FormValue("url"); remoteURL != "" {
			opts.URL = remoteURL
		}
		if expiry := r.FormValue("expires"); expiry != "" {
			if d, err := time.ParseDuration(expiry); err == nil {
				opts.Expires = d
			}
		}
	}

	if maxDown := r.Header.Get("Max-Downloads"); maxDown != "" {
		if n, err := strconv.Atoi(maxDown); err == nil && n > 0 {
			opts.MaxDownloads = n
		}
	}

	ct := r.Header.Get("Content-Type")

	if strings.HasPrefix(ct, "multipart/form-data") {
		name, ct, size, body, err := h.parseMultipart(r)

		return name, ct, size, body, opts, err
	}

	if len(ct) > 0 && ct != "application/octet-stream" && ct != "multipart/form-data" {
		filename = "file"
		contentType = ct
		size = r.ContentLength
		body = r.Body

		return filename, contentType, size, body, opts, nil
	}

	filename = "file"
	contentType = "application/octet-stream"
	size = r.ContentLength
	body = r.Body

	return filename, contentType, size, body, opts, nil
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

func (h *UploadHandler) fetchRemote(remoteURL string) (filename, contentType string, size int64, body io.Reader, err error) {
	if err := netsec.IsPrivateURL(remoteURL); err != nil {
		return "", "", 0, nil, fmt.Errorf("url check failed: %w", err)
	}

	resp, err := h.httpClient.Get(remoteURL)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, nil, fmt.Errorf("remote returned: %s", resp.Status)
	}

	contentType = resp.Header.Get("Content-Type")
	size = resp.ContentLength

	cd := resp.Header.Get("Content-Disposition")
	filename = fetch.FilenameFromHeader(cd, remoteURL)
	if filename == "" {
		filename = "file"
	}

	return filename, contentType, size, resp.Body, nil
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
