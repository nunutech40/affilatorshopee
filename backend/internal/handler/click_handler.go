package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type ClickHandler struct{ clicks *repository.ClickRepository }

func NewClickHandler(clicks *repository.ClickRepository) *ClickHandler {
	return &ClickHandler{clicks: clicks}
}

func (h *ClickHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FILE", "File CSV tidak valid atau terlalu besar")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "FILE_REQUIRED", "Pilih file CSV terlebih dahulu")
		return
	}
	defer file.Close()
	reader := csv.NewReader(io.LimitReader(file, 2<<20))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CSV", "CSV kosong atau tidak dapat dibaca")
		return
	}
	indexes := map[string]int{}
	for index, value := range header {
		indexes[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")))] = index
	}
	for _, key := range []string{"klik id", "waktu klik", "tag_link"} {
		if _, ok := indexes[key]; !ok {
			writeError(w, http.StatusBadRequest, "INVALID_CSV", fmt.Sprintf("Kolom %s tidak ditemukan", key))
			return
		}
	}
	events := []model.ClickEvent{}
	for row := 2; ; row++ {
		values, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_CSV", fmt.Sprintf("Baris %d tidak valid", row))
			return
		}
		get := func(key string) string {
			index := indexes[key]
			if index >= len(values) {
				return ""
			}
			return strings.TrimSpace(values[index])
		}
		clickedAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", get("waktu klik"), time.Local)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_CSV", fmt.Sprintf("Waktu klik pada baris %d tidak valid", row))
			return
		}
		events = append(events, model.ClickEvent{ClickID: get("klik id"), ClickedAt: clickedAt, Region: get("wilayah klik"), TrackingTag: get("tag_link"), Referrer: get("perujuk")})
	}
	result, err := h.clicks.Sync(r.Context(), events)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyinkronkan data klik")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
