package handler

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type NicheHandler struct{ niches *repository.NicheRepository }

func NewNicheHandler(niches *repository.NicheRepository) *NicheHandler {
	return &NicheHandler{niches: niches}
}

func (h *NicheHandler) List(w http.ResponseWriter, r *http.Request) {
	niches, err := h.niches.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil niche")
		return
	}
	writeJSON(w, http.StatusOK, niches)
}

func (h *NicheHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if strings.TrimSpace(request.Name) == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Nama niche wajib diisi")
		return
	}
	niche, err := h.niches.Create(r.Context(), request.Name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeError(w, http.StatusConflict, "DUPLICATE", "Niche tersebut sudah ada")
			return
		}
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menambah niche")
		return
	}
	writeJSON(w, http.StatusCreated, niche)
}

func (h *NicheHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.niches.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menghapus niche")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NicheHandler) ReplaceProduct(w http.ResponseWriter, r *http.Request) {
	var request struct {
		NicheIDs []string `json:"niche_ids"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if err := h.niches.ReplaceProductNiches(r.Context(), chi.URLParam(r, "id"), request.NicheIDs); err != nil {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Produk atau niche tidak valid")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"saved": true})
}
