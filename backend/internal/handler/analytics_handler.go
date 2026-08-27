package handler

import (
	"database/sql"
	"net/http"

	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type AnalyticsHandler struct {
	db          *sql.DB
	clicks      *repository.ClickRepository
	commissions *repository.CommissionRepository
	repo        *repository.ProductRepository
}

func NewAnalyticsHandler(db *sql.DB, clicks *repository.ClickRepository, commissions *repository.CommissionRepository, repo *repository.ProductRepository) *AnalyticsHandler {
	return &AnalyticsHandler{db: db, clicks: clicks, commissions: commissions, repo: repo}
}

// Reset truncates click/commission events and resets product counters. Idempotent sync already prevents double-count on re-upload via event_id PK.
func (h *AnalyticsHandler) Reset(w http.ResponseWriter, r *http.Request) {
	if _, err := h.db.ExecContext(r.Context(), `TRUNCATE click_events, commission_events`); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal reset click/commission")
		return
	}
	if _, err := h.db.ExecContext(r.Context(), `UPDATE products SET click_count=0, last_clicked_at=NULL, sales_count=0, pending_sales_count=0, commission_total=0`); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal reset product counters")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset ok"})
}
