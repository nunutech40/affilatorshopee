package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type ContentHandler struct{ repo *repository.ContentRepository }

func NewContentHandler(repo *repository.ContentRepository) *ContentHandler {
	return &ContentHandler{repo: repo}
}

func (h *ContentHandler) ListNiches(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListNiches(r.Context())
	if err != nil {
		writeError(w, 500, "DATABASE_ERROR", "Gagal mengambil niche konten")
		return
	}
	writeJSON(w, 200, items)
}
func (h *ContentHandler) CreateNiche(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(w, r, &b); err != nil || strings.TrimSpace(b.Name) == "" {
		writeError(w, 400, "VALIDATION_ERROR", "Nama niche konten wajib diisi")
		return
	}
	item, err := h.repo.CreateNiche(r.Context(), b.Name)
	if err != nil {
		writeError(w, 409, "DUPLICATE", "Niche konten sudah ada")
		return
	}
	writeJSON(w, 201, item)
}
func (h *ContentHandler) DeleteNiche(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteNiche(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, 500, "DATABASE_ERROR", "Gagal mengarsipkan niche konten")
		return
	}
	w.WriteHeader(204)
}
func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	items, total, err := h.repo.List(r.Context(), r.URL.Query().Get("content_niche_id"), r.URL.Query().Get("platform"), r.URL.Query().Get("status"), r.URL.Query().Get("search"), page, limit)
	if err != nil {
		writeError(w, 500, "DATABASE_ERROR", "Gagal mengambil bank konten")
		return
	}
	writeJSON(w, 200, map[string]interface{}{"items": items, "total": total, "page": page, "limit": limit})
}
func (h *ContentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Platform       string              `json:"platform"`
		ExternalPostID string              `json:"external_post_id"`
		CanonicalURL   string              `json:"canonical_url"`
		AuthorHandle   string              `json:"author_handle"`
		OriginalText   string              `json:"original_text"`
		Media          []string            `json:"media"`
		SourceQuery    string              `json:"source_query"`
		PublishedAt    *time.Time          `json:"published_at"`
		Status         string              `json:"status"`
		NicheIDs       []string            `json:"niche_ids"`
		ProductTypeIDs []string            `json:"product_type_ids"`
		Stats          *model.ContentStats `json:"stats"`
	}
	if err := decodeJSON(w, r, &b); err != nil {
		writeError(w, 400, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	if strings.TrimSpace(b.OriginalText) == "" || strings.TrimSpace(b.CanonicalURL) == "" {
		writeError(w, 400, "VALIDATION_ERROR", "URL dan konten asli wajib diisi")
		return
	}
	if b.Platform == "" {
		b.Platform = "x"
	}
	if b.Status == "" {
		b.Status = "discovered"
	}
	item, err := h.repo.Create(r.Context(), model.ContentItem{Platform: b.Platform, ExternalPostID: b.ExternalPostID, CanonicalURL: b.CanonicalURL, AuthorHandle: b.AuthorHandle, OriginalText: b.OriginalText, Media: b.Media, PublishedAt: b.PublishedAt, SourceQuery: b.SourceQuery, Status: b.Status}, b.NicheIDs, b.ProductTypeIDs, b.Stats)
	if err != nil {
		writeError(w, 400, "CREATE_ERROR", "Konten gagal disimpan: "+err.Error())
		return
	}
	writeJSON(w, 201, item)
}
