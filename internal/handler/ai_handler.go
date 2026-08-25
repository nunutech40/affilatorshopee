package handler

import (
	"errors"
	"net/http"

	"github.com/nunutech40/affilatorshopee/internal/service"
)

type AIHandler struct {
	products  *service.ProductService
	ai        *service.AIService
	semaphore chan struct{}
}

func NewAIHandler(products *service.ProductService, ai *service.AIService) *AIHandler {
	return &AIHandler{products: products, ai: ai, semaphore: make(chan struct{}, 1)}
}

type reformatRequest struct {
	ProductIDs []string `json:"product_ids"`
	Model      *string  `json:"model"`
}

func (h *AIHandler) Models(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, service.ListModels())
}

func (h *AIHandler) Reformat(w http.ResponseWriter, r *http.Request) {
	select {
	case h.semaphore <- struct{}{}:
		defer func() { <-h.semaphore }()
	default:
		writeError(w, http.StatusTooManyRequests, "AI_BUSY", "Request AI lain sedang diproses")
		return
	}
	var request reformatRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	model := ""
	if request.Model != nil {
		model = *request.Model
	}
	summary, err := h.products.Reformat(r.Context(), request.ProductIDs, h.ai, model)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, "AI_PROVIDER_ERROR", "Gagal memproses reformat AI")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}
