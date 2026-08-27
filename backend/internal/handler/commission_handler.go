package handler

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type CommissionHandler struct {
	commissions *repository.CommissionRepository
}

func NewCommissionHandler(commissions *repository.CommissionRepository) *CommissionHandler {
	return &CommissionHandler{commissions: commissions}
}

func (h *CommissionHandler) ListSold(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 20
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			offset = (n - 1) * limit
		}
	}
	search := q.Get("search")
	items, total, err := h.commissions.ListSoldProducts(r.Context(), limit, offset, search)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk terjual")
		return
	}
	page := 1
	if offset > 0 {
		page = offset/limit + 1
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *CommissionHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FILE", "File CSV tidak valid atau terlalu besar")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "FILE_REQUIRED", "Pilih file CSV terlebih dahulu")
		return
	}
	defer file.Close()
	reader := csv.NewReader(io.LimitReader(file, 4<<20))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_CSV", "CSV kosong atau tidak dapat dibaca")
		return
	}
	indexes := map[string]int{}
	for i, v := range header {
		indexes[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "\ufeff")))] = i
	}
	// Required columns - flexible naming
	has := func(keys ...string) bool {
		for _, k := range keys {
			if _, ok := indexes[k]; ok {
				return true
			}
		}
		return false
	}
	if !has("event id", "event_id") {
		writeError(w, http.StatusBadRequest, "INVALID_CSV", "Kolom event_id tidak ditemukan")
		return
	}
	events := []model.CommissionEvent{}
	for row := 2; ; row++ {
		values, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_CSV", fmt.Sprintf("Baris %d tidak valid", row))
			return
		}
		get := func(keys ...string) string {
			for _, k := range keys {
				if idx, ok := indexes[k]; ok && idx < len(values) {
					if v := strings.TrimSpace(values[idx]); v != "" {
						return v
					}
				}
			}
			return ""
		}
		eventID := get("event id", "event_id")
		if eventID == "" {
			continue
		}
		trackingTag := get("tag_link", "tracking tag", "tracking_tag", "tag link")
		orderedAtStr := get("ordered at", "ordered_at", "waktu pesan", "waktu pesanan")
		var orderedAt *time.Time
		if orderedAtStr != "" {
			for _, layout := range []string{"2006-01-02 15:04:05", "2006/01/02 15:04:05", time.RFC3339, "2006-01-02"} {
				if t, err := time.ParseInLocation(layout, orderedAtStr, time.Local); err == nil {
					orderedAt = &t
					break
				}
			}
		}
		quantity := 1
		if qStr := get("quantity", "qty", "jumlah"); qStr != "" {
			if q, err := strconv.Atoi(strings.ReplaceAll(qStr, ",", "")); err == nil {
				quantity = q
			}
		}
		commission := int64(0)
		if cStr := get("commission total", "commission_total", "komisi", "total komisi"); cStr != "" {
			clean := strings.ReplaceAll(strings.ReplaceAll(cStr, ",", ""), ".", "")
			// try as integer cents - if contains decimal, parse float
			if v, err := strconv.ParseFloat(strings.ReplaceAll(cStr, ",", ""), 64); err == nil {
				commission = int64(v)
			} else if v, err := strconv.ParseInt(clean, 10, 64); err == nil {
				commission = v
			}
		}
		events = append(events, model.CommissionEvent{
			EventID:         eventID,
			OrderID:         get("order id", "order_id", "id pesanan"),
			ItemID:          get("item id", "item_id"),
			ModelID:         get("model id", "model_id"),
			OrderStatus:     get("order status", "order_status", "status pesanan", "status"),
			OrderedAt:       orderedAt,
			TrackingTag:     trackingTag,
			Quantity:        quantity,
			CommissionTotal: commission,
		})
	}
	result, err := h.commissions.Sync(r.Context(), events)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyinkronkan komisi")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
