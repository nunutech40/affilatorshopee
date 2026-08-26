package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type PostLogHandler struct {
	logs *service.PostLogService
}

func NewPostLogHandler(logs *service.PostLogService) *PostLogHandler {
	return &PostLogHandler{logs: logs}
}

func (h *PostLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	var log model.PostLog
	if err := decodeJSON(w, r, &log); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if err := h.logs.Create(r.Context(), &log); err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyimpan riwayat posting")
		return
	}
	writeJSON(w, http.StatusCreated, log)
}

func (h *PostLogHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.logs.ListByProduct(r.Context(), chi.URLParam(r, "productID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil riwayat posting")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *PostLogHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	var items []model.PostLog
	var err error
	if productID := r.URL.Query().Get("product_id"); productID != "" {
		items, err = h.logs.ListByProduct(r.Context(), productID)
	} else {
		items, err = h.logs.ListAll(r.Context())
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil riwayat posting")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "page": 1, "limit": len(items), "total": len(items)})
}
