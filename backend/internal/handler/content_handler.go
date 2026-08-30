package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type ContentHandler struct {
	repo *repository.ContentRepository
	ai   *service.AIService
}

func NewContentHandler(repo *repository.ContentRepository, ai ...*service.AIService) *ContentHandler {
	h := &ContentHandler{repo: repo}
	if len(ai) > 0 {
		h.ai = ai[0]
	}
	return h
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
		Platform        string              `json:"platform"`
		ContentFormat   string              `json:"content_format"`
		ThreadPostCount int                 `json:"thread_post_count"`
		ExternalPostID  string              `json:"external_post_id"`
		CanonicalURL    string              `json:"canonical_url"`
		AuthorHandle    string              `json:"author_handle"`
		OriginalText    string              `json:"original_text"`
		Media           []string            `json:"media"`
		SourceQuery     string              `json:"source_query"`
		PublishedAt     *time.Time          `json:"published_at"`
		Status          string              `json:"status"`
		NicheIDs        []string            `json:"niche_ids"`
		ProductTypeIDs  []string            `json:"product_type_ids"`
		Stats           *model.ContentStats `json:"stats"`
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
	if b.ContentFormat == "" {
		if b.ThreadPostCount > 1 {
			b.ContentFormat = "thread"
		} else {
			b.ContentFormat = "post"
		}
	}
	if b.Status == "" {
		b.Status = "discovered"
	}
	item, err := h.repo.Create(r.Context(), model.ContentItem{Platform: b.Platform, ContentFormat: b.ContentFormat, ExternalPostID: b.ExternalPostID, CanonicalURL: b.CanonicalURL, AuthorHandle: b.AuthorHandle, OriginalText: b.OriginalText, Media: b.Media, PublishedAt: b.PublishedAt, SourceQuery: b.SourceQuery, Status: b.Status}, b.NicheIDs, b.ProductTypeIDs, b.Stats)
	if err != nil {
		writeError(w, 400, "CREATE_ERROR", "Konten gagal disimpan: "+err.Error())
		return
	}
	writeJSON(w, 201, item)
}

type contentPayload struct {
	Platform       string              `json:"platform"`
	ContentFormat  string              `json:"content_format"`
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

func (h *ContentHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.repo.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Konten tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 500, "DATABASE_ERROR", "Gagal mengambil detail konten")
		return
	}
	writeJSON(w, 200, item)
}

func (h *ContentHandler) Patch(w http.ResponseWriter, r *http.Request) {
	var b contentPayload
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
	if b.ContentFormat == "" {
		b.ContentFormat = "post"
	}
	item, err := h.repo.Update(r.Context(), chi.URLParam(r, "id"), repository.ContentUpdate{Platform: b.Platform, ContentFormat: b.ContentFormat, ExternalPostID: b.ExternalPostID, CanonicalURL: b.CanonicalURL, AuthorHandle: b.AuthorHandle, OriginalText: b.OriginalText, Media: b.Media, PublishedAt: b.PublishedAt, SourceQuery: b.SourceQuery, Status: b.Status}, b.NicheIDs, b.ProductTypeIDs, b.Stats)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Konten tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 400, "UPDATE_ERROR", "Konten gagal diperbarui: "+err.Error())
		return
	}
	writeJSON(w, 200, item)
}

func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.repo.Delete(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Konten tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 500, "DELETE_ERROR", "Konten gagal dihapus")
		return
	}
	w.WriteHeader(204)
}

type variantPayload struct {
	Name     string `json:"name"`
	Text     string `json:"text"`
	Source   string `json:"source"`
	Model    string `json:"model"`
	Position int    `json:"position"`
}

func (h *ContentHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	var b variantPayload
	if err := decodeJSON(w, r, &b); err != nil {
		writeError(w, 400, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	v, err := h.repo.CreateVariant(r.Context(), chi.URLParam(r, "id"), model.ContentVariant{Name: b.Name, Text: b.Text, Source: b.Source, Model: b.Model, Position: b.Position})
	if err != nil {
		writeError(w, 400, "CREATE_ERROR", "Varian gagal disimpan: "+err.Error())
		return
	}
	writeJSON(w, 201, v)
}
func (h *ContentHandler) PatchVariant(w http.ResponseWriter, r *http.Request) {
	var b variantPayload
	if err := decodeJSON(w, r, &b); err != nil {
		writeError(w, 400, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	v, err := h.repo.UpdateVariant(r.Context(), chi.URLParam(r, "variantID"), model.ContentVariant{Name: b.Name, Text: b.Text, Source: b.Source, Model: b.Model, Position: b.Position})
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Varian tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 400, "UPDATE_ERROR", "Varian gagal diperbarui: "+err.Error())
		return
	}
	writeJSON(w, 200, v)
}
func (h *ContentHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	err := h.repo.DeleteVariant(r.Context(), chi.URLParam(r, "variantID"))
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, 404, "NOT_FOUND", "Varian tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, 500, "DELETE_ERROR", "Varian gagal dihapus")
		return
	}
	w.WriteHeader(204)
}

func (h *ContentHandler) ReformatVariant(w http.ResponseWriter, r *http.Request) {
	if h.ai == nil {
		writeError(w, 503, "AI_UNAVAILABLE", "AI belum dikonfigurasi")
		return
	}
	var b struct {
		Model string `json:"model"`
		Name  string `json:"name"`
	}
	if err := decodeJSON(w, r, &b); err != nil {
		writeError(w, 400, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	item, err := h.repo.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 404, "NOT_FOUND", "Konten tidak ditemukan")
		return
	}
	cm := "trending"
	p := model.Product{ID: item.ID, RawText: item.OriginalText, ContentFormat: item.ContentFormat, ContentModel: &cm}
	results, err := h.ai.ReformatContent(r.Context(), []model.Product{p}, b.Model)
	if err != nil {
		writeError(w, 502, "AI_PROVIDER_ERROR", err.Error())
		return
	}
	if len(results) == 0 || strings.TrimSpace(results[0].ContentText) == "" {
		writeError(w, 502, "AI_EMPTY", "AI tidak mengembalikan varian")
		return
	}
	v, err := h.repo.CreateVariant(r.Context(), item.ID, model.ContentVariant{Name: b.Name, Text: results[0].ContentText, Source: "ai", Model: b.Model})
	if err != nil {
		writeError(w, 500, "CREATE_ERROR", "Varian AI gagal disimpan")
		return
	}
	writeJSON(w, 201, v)
}

func (h *ContentHandler) CleanRaw(w http.ResponseWriter, r *http.Request) {
	if h.ai == nil {
		writeError(w, 503, "AI_UNAVAILABLE", "AI belum dikonfigurasi")
		return
	}
	var b struct {
		Model string `json:"model"`
	}
	if err := decodeJSON(w, r, &b); err != nil {
		writeError(w, 400, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	item, err := h.repo.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 404, "NOT_FOUND", "Konten tidak ditemukan")
		return
	}
	cm := "trending"
	results, err := h.ai.CleanRaw(r.Context(), []model.Product{{ID: item.ID, RawText: item.OriginalText, ContentModel: &cm}}, b.Model)
	if err != nil {
		writeError(w, 502, "AI_PROVIDER_ERROR", err.Error())
		return
	}
	if len(results) == 0 || strings.TrimSpace(results[0].CleanedRawText) == "" {
		writeError(w, 502, "AI_EMPTY", "AI tidak mengembalikan raw bersih")
		return
	}
	nicheIDs := make([]string, 0, len(item.Niches))
	for _, n := range item.Niches {
		nicheIDs = append(nicheIDs, n.ID)
	}
	typeIDs := make([]string, 0, len(item.ProductTypes))
	for _, n := range item.ProductTypes {
		typeIDs = append(typeIDs, n.ID)
	}
	updated, err := h.repo.Update(r.Context(), item.ID, repository.ContentUpdate{Platform: item.Platform, ExternalPostID: item.ExternalPostID, CanonicalURL: item.CanonicalURL, AuthorHandle: item.AuthorHandle, OriginalText: item.OriginalText, CleanedOriginalText: results[0].CleanedRawText, Media: item.Media, PublishedAt: item.PublishedAt, SourceQuery: item.SourceQuery, Status: item.Status}, nicheIDs, typeIDs, nil)
	if err != nil {
		writeError(w, 500, "UPDATE_ERROR", "Raw bersih gagal disimpan: "+err.Error())
		return
	}
	writeJSON(w, 200, updated)
}
