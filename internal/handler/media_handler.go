package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type MediaHandler struct{ media *service.MediaService }

func NewMediaHandler(media *service.MediaService) *MediaHandler { return &MediaHandler{media: media} }

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.media.List(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil media")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *MediaHandler) Download(w http.ResponseWriter, r *http.Request) {
	archive, err := h.media.Zip(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Gagal membuat ZIP media")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="product-media.zip"`)
	_, _ = w.Write(archive.Bytes())
}
