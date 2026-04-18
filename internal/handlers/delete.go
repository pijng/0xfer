package handlers

import (
	"net/http"

	"0xfer/internal/services"
)

type DeleteHandler struct {
	service *services.FileService
}

func NewDeleteHandler(service *services.FileService) http.Handler {
	h := &DeleteHandler{service: service}

	return http.HandlerFunc(h.serve)
}

func (h *DeleteHandler) serve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	secret := r.PathValue("secret")
	if id == "" || secret == "" {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	if err := h.service.Delete(r.Context(), id, secret); err != nil {
		http.Error(w, "not found", http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}