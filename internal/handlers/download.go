package handlers

import (
	"io"
	"net/http"
	"strconv"

	"0xfer/internal/services"
)

type DownloadHandler struct {
	service *services.FileService
}

func NewDownloadHandler(service *services.FileService) http.Handler {
	h := &DownloadHandler{service: service}

	return http.HandlerFunc(h.serve)
}

func (h *DownloadHandler) serve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	f, file, err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}
	defer func() { _ = file.Close() }()

	w.Header().Set("Content-Type", f.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	w.Header().Set("Content-Disposition", `attachment; filename="`+escapeQuotes(f.Filename)+`"`)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Accept-Ranges", "bytes")

	_, _ = io.Copy(w, file)
}

func escapeQuotes(s string) string {
	result := make([]byte, 0, len(s))
	for _, c := range s {
		if c == '"' {
			result = append(result, '\\', '"')
		} else {
			result = append(result, byte(c))
		}
	}

	return string(result)
}
