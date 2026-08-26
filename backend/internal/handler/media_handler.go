package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type MediaHandler struct{ media *service.MediaService }

type addMediaRequest struct {
	ImageURLs []string `json:"image_urls"`
	VideoURL  *string  `json:"video_url"`
}

func NewMediaHandler(media *service.MediaService) *MediaHandler { return &MediaHandler{media: media} }

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.media.List(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil media")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *MediaHandler) Add(w http.ResponseWriter, r *http.Request) {
	var request addMediaRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if len(request.ImageURLs) == 0 && (request.VideoURL == nil || strings.TrimSpace(*request.VideoURL) == "") {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Minimal satu URL image atau video wajib diisi")
		return
	}
	result := h.media.DownloadProductMedia(r.Context(), chi.URLParam(r, "id"), request.ImageURLs, request.VideoURL)
	if len(result.Downloaded) == 0 && len(result.Failed) > 0 {
		writeJSON(w, http.StatusBadRequest, result)
		return
	}
	writeJSON(w, http.StatusCreated, result)
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

func (h *MediaHandler) File(w http.ResponseWriter, r *http.Request) {
	item, file, err := h.media.Open(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "mediaID"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Media tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Gagal membaca media")
		return
	}
	defer file.Close()
	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Content-Disposition", `inline; filename="`+strings.ReplaceAll(item.Filename, `"`, "")+`"`)
	http.ServeContent(w, r, item.Filename, item.CreatedAt, file)
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.media.Remove(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "mediaID"))
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Media tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Gagal menghapus media")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
