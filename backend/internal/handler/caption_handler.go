package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type CaptionHandler struct {
	products   *service.ProductService
	captions   *service.CaptionService
	variations *service.CaptionVariationService
}

func NewCaptionHandler(products *service.ProductService, captions *service.CaptionService, variations *service.CaptionVariationService) *CaptionHandler {
	return &CaptionHandler{products: products, captions: captions, variations: variations}
}

type generateCaptionRequest struct {
	ProductID string   `json:"product_id"`
	Template  string   `json:"template"`
	Hashtags  []string `json:"hashtags"`
}

func (h *CaptionHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var request generateCaptionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	product, err := h.products.GetByID(r.Context(), request.ProductID)
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk")
		return
	}
	caption, err := h.captions.Generate(product, request.Template, request.Hashtags)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "CAPTION_ERROR", "Gagal membuat caption")
		return
	}
	writeJSON(w, http.StatusOK, caption)
}

type generateVariationsRequest struct {
	ProductID string   `json:"product_id"`
	Template  string   `json:"template"`
	Count     int      `json:"count"`
	Hashtags  []string `json:"hashtags"`
}

func (h *CaptionHandler) GenerateVariations(w http.ResponseWriter, r *http.Request) {
	var request generateVariationsRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	product, err := h.products.GetByID(r.Context(), request.ProductID)
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk")
		return
	}
	generated, err := h.captions.GenerateVariations(product, request.Template, request.Count, request.Hashtags)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "CAPTION_ERROR", "Gagal membuat variasi")
		return
	}
	result := make([]model.CaptionVariation, 0, len(generated))
	for i, item := range generated {
		variation := model.CaptionVariation{ProductID: product.ID, Label: "Variation " + string(rune('A'+i)), Template: item.Template, Caption: item.Caption, Hashtags: item.Hashtags}
		if err := h.variations.Create(r.Context(), &variation); err != nil {
			writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyimpan variasi")
			return
		}
		result = append(result, variation)
	}
	writeJSON(w, http.StatusCreated, result)
}

func (h *CaptionHandler) ListVariations(w http.ResponseWriter, r *http.Request) {
	items, err := h.variations.ListByProduct(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil variasi")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *CaptionHandler) PatchVariation(w http.ResponseWriter, r *http.Request) {
	variation, err := h.variations.GetByID(r.Context(), chi.URLParam(r, "variationID"))
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Variasi tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil variasi")
		return
	}
	var patch struct {
		Label    *string   `json:"label"`
		Caption  *string   `json:"caption"`
		Hashtags *[]string `json:"hashtags"`
	}
	if err := decodeJSON(w, r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if patch.Label != nil {
		variation.Label = *patch.Label
	}
	if patch.Caption != nil {
		variation.Caption = *patch.Caption
	}
	if patch.Hashtags != nil {
		variation.Hashtags = *patch.Hashtags
	}
	if err := h.variations.Update(r.Context(), variation); err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengubah variasi")
		return
	}
	updated, _ := h.variations.GetByID(r.Context(), variation.ID)
	writeJSON(w, http.StatusOK, updated)
}

func (h *CaptionHandler) DeleteVariation(w http.ResponseWriter, r *http.Request) {
	if err := h.variations.Delete(r.Context(), chi.URLParam(r, "variationID")); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menghapus variasi")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
