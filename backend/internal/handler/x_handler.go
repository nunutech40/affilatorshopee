package handler

import (
	"errors"
	"net/http"

	"github.com/nunutech40/affilatorshopee/internal/service"
)

type XHandler struct {
	xService *service.XService
}

func NewXHandler(xService *service.XService) *XHandler {
	return &XHandler{xService: xService}
}

type importXRequest struct {
	XURL         string  `json:"x_url"`
	ContentModel *string `json:"content_model"`
}

func (h *XHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req importXRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	product, media, err := h.xService.ImportFromX(r.Context(), req.XURL, req.ContentModel)
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, "X_IMPORT_ERROR", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"product": product,
		"media":   media,
	})
}
