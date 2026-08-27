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

func parseDateFilter(values map[string]string) (start, end *string) {
	if m := values["month"]; m != "" {
		if t, err := time.Parse("2006-01", m); err == nil {
			s := t.Format("2006-01-02")
			e := t.AddDate(0, 1, -1).Format("2006-01-02") + " 23:59:59"
			return &s, &e
		}
	}
	var s, e *string
	if v := values["start_date"]; v != "" {
		tmp := v
		s = &tmp
	}
	if v := values["end_date"]; v != "" {
		if !strings.Contains(v, " ") {
			tmp := v + " 23:59:59"
			e = &tmp
		} else {
			tmp := v
			e = &tmp
		}
	}
	return s, e
}

func (h *CommissionHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 20
	offset := 0
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			offset = (n - 1) * limit
		}
	}
	search := q.Get("search")
	vals := map[string]string{"start_date": q.Get("start_date"), "end_date": q.Get("end_date"), "month": q.Get("month")}
	start, end := parseDateFilter(vals)
	items, total, err := h.commissions.ListEvents(r.Context(), limit, offset, search, start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil detail komisi")
		return
	}
	page := 1
	if offset > 0 {
		page = offset/limit + 1
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "total": total, "page": page, "limit": limit})
}

func (h *CommissionHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	vals := map[string]string{"start_date": q.Get("start_date"), "end_date": q.Get("end_date"), "month": q.Get("month")}
	start, end := parseDateFilter(vals)
	summary, err := h.commissions.GetSummary(r.Context(), q.Get("search"), start, end)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil ringkasan komisi")
		return
	}
	writeJSON(w, http.StatusOK, summary)
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
	vals2 := map[string]string{"start_date": q.Get("start_date"), "end_date": q.Get("end_date"), "month": q.Get("month")}
	start, end := parseDateFilter(vals2)
	items, total, err := h.commissions.ListSoldProducts(r.Context(), limit, offset, search, start, end)
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
	// Required columns - flexible naming (support Shopee ID Pemesanan as event_id)
	has := func(keys ...string) bool {
		for _, k := range keys {
			if _, ok := indexes[k]; ok {
				return true
			}
		}
		return false
	}
	if !has("event id", "event_id", "id pemesanan", "id pesanan") {
		writeError(w, http.StatusBadRequest, "INVALID_CSV", "Kolom ID Pemesanan / event_id tidak ditemukan")
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
		orderID := get("order id", "order_id", "id pemesanan", "id pesanan", "kode pesanan affiliate")
		itemID := get("item id", "item_id", "id barang", "product id", "itemid")
		modelID := get("model id", "model_id", "id model")
		eventID := get("event id", "event_id")
		trackingTag := get("tag_link1", "tag_link2", "tag_link3", "tag_link4", "tag_link5", "tag_link", "tracking tag", "tracking_tag", "tag link")
		// Shopee can report multiple products/models under one order. Use the
		// composite identity so re-imports update the same line without collisions.
		if orderID != "" {
			eventID = orderID + "|" + itemID + "|" + modelID
		} else if eventID == "" {
			continue
		}
		orderedAtStr := get("ordered at", "ordered_at", "waktu pemesanan", "waktu pesan", "waktu pesanan")
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
		if qStr := get("jumlah", "quantity", "qty"); qStr != "" {
			if q, err := strconv.Atoi(strings.ReplaceAll(qStr, ",", "")); err == nil {
				quantity = q
			}
		}
		commission := int64(0)
		// support both Indonesian and English headers, with (Rp) suffix - prioritize bersih
		if cStr := get("komisi bersih affiliate (rp)", "komisi bersih affiliate(rp)", "total komisi per produk(rp)", "total komisi per pesanan(rp)", "komisi barang shopee(rp)", "commission total", "commission_total", "komisi", "total komisi"); cStr != "" {
			// handle both "2.546,18" (ID) and "2546.18" (EN)
			cleanID := strings.ReplaceAll(cStr, ".", "")
			cleanID = strings.ReplaceAll(cleanID, ",", ".")
			cleanEN := strings.ReplaceAll(cStr, ",", "")
			var v float64
			var err error
			// try EN first (dot decimal), then ID
			if v, err = strconv.ParseFloat(cleanEN, 64); err == nil {
				// if original had ',' as decimal, EN will be wrong (e.g. "2.546,18" -> "2546.18" after remove ","? Actually cleanEN "2.546.18" -> 2.546)
				// detect ID format: contains ',' and '.' -> use ID parse
				if strings.Contains(cStr, ",") && strings.Contains(cStr, ".") {
					if v2, err2 := strconv.ParseFloat(cleanID, 64); err2 == nil {
						v = v2
					}
				}
				commission = int64(v)
			} else if v, err := strconv.ParseFloat(cleanID, 64); err == nil {
				commission = int64(v)
			}
		}
		// allow empty tracking tag - still record as sold
		events = append(events, model.CommissionEvent{
			EventID:         eventID,
			OrderID:         orderID,
			ItemID:          itemID,
			ModelID:         modelID,
			OrderStatus:     get("order status", "order_status", "status pesanan", "status", "status produk affiliate"),
			OrderedAt:       orderedAt,
			TrackingTag:     trackingTag,
			Quantity:        quantity,
			CommissionTotal: commission,
			ItemName:        get("nama barange", "nama barang", "item name", "item_name", "product name", "product_name", "nama produk", "nama item"),
			ShopName:        get("nama toko", "shop name", "shop_name", "nama toko", "toko"),
		})
	}
	result, err := h.commissions.Sync(r.Context(), events)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyinkronkan komisi")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
