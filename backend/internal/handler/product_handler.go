package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
	"github.com/nunutech40/affilatorshopee/internal/service"
)

type ProductHandler struct {
	products *service.ProductService
	media    *service.MediaService
}

func NewProductHandler(products *service.ProductService, media *service.MediaService) *ProductHandler {
	return &ProductHandler{products: products, media: media}
}

type createProductRequest struct {
	RawText        string   `json:"raw_text"`
	ProductName    *string  `json:"product_name"`
	ShopeeLink     string   `json:"shopee_link"`
	TrackingTag    *string  `json:"tracking_tag"`
	SourceCategory string   `json:"source_category"`
	ImageURL       *string  `json:"image_url"`
	ImageURLs      []string `json:"image_urls"`
	VideoURL       *string  `json:"video_url"`
	ContentModel   *string  `json:"content_model"`
	Notes          *string  `json:"notes"`
}

type productPatch struct {
	ReformattedText  *string   `json:"reformatted_text"`
	ResetReformatted bool      `json:"reset_reformatted"`
	ProductName      *string   `json:"product_name"`
	ShopeeLink       *string   `json:"shopee_link"`
	TrackingTag      *string   `json:"tracking_tag"`
	ImageURL         **string  `json:"image_url"`
	ImageURLs        *[]string `json:"image_urls"`
	VideoURL         **string  `json:"video_url"`
	NormalPrice      **int     `json:"normal_price"`
	SalePrice        **int     `json:"sale_price"`
	DiscountPercent  **int     `json:"discount_percent"`
	Rating           **float64 `json:"rating"`
	SoldCount        **string  `json:"sold_count"`
	ReviewCount      **string  `json:"review_count"`
	Keyword          **string  `json:"keyword"`
	Problem          **string  `json:"problem"`
	Cluster          **string  `json:"cluster"`
	ContentModel     **string  `json:"content_model"`
	CaptureAngle     **string  `json:"capture_angle"`
	Benefit1         **string  `json:"benefit_1"`
	Benefit2         **string  `json:"benefit_2"`
	Benefit3         **string  `json:"benefit_3"`
	Urgency          **string  `json:"urgency"`
	CaptionTemplate  *string   `json:"caption_template"`
	HashtagPool      *[]string `json:"hashtag_pool"`
	Notes            **string  `json:"notes"`
	SourceCategory   *string   `json:"source_category"`
	Status           *string   `json:"status"`
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := parsePositive(query.Get("page"), 1)
	limit := parsePositive(query.Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	items, total, err := h.products.List(r.Context(), repository.ProductListFilter{
		Cluster: query.Get("cluster"), ContentModel: query.Get("content_model"), SourceCategory: query.Get("source_category"),
		Status: query.Get("status"), Search: query.Get("search"), Page: page, Limit: limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": items, "page": page, "limit": limit, "total": total})
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createProductRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	product := &model.Product{
		RawText: request.RawText, ProductName: request.ProductName, ShopeeLink: request.ShopeeLink, TrackingTag: "", SourceCategory: request.SourceCategory, ImageURL: request.ImageURL,
		ImageURLs: request.ImageURLs, VideoURL: request.VideoURL, ContentModel: request.ContentModel, Notes: request.Notes,
	}
	if request.TrackingTag != nil {
		product.TrackingTag = *request.TrackingTag
	}
	if product.SourceCategory == "" {
		product.SourceCategory = "raw_text"
	}
	if len(product.ImageURLs) == 0 && product.ImageURL != nil && *product.ImageURL != "" {
		product.ImageURLs = []string{*product.ImageURL}
	}
	if len(product.ImageURLs) > 0 {
		product.ImageURL = &product.ImageURLs[0]
	}
	if err := h.products.Create(r.Context(), product); err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menyimpan produk")
		return
	}
	mediaResult := service.MediaDownloadSummary{Downloaded: []model.MediaFile{}, Failed: []service.MediaDownloadFailure{}}
	if h.media != nil {
		mediaResult = h.media.DownloadProductMedia(r.Context(), product.ID, product.ImageURLs, product.VideoURL)
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"product": product, "media": mediaResult})
}

func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	product, err := h.products.GetByID(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk")
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	product, err := h.products.GetByID(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Produk tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengambil produk")
		return
	}
	var patch productPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Body JSON tidak valid")
		return
	}
	applyProductPatch(product, patch)
	if err := h.products.Update(r.Context(), product); err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Produk tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal mengubah produk")
		return
	}
	updated, err := h.products.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal membaca produk setelah update")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.products.Delete(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, service.ErrValidation) {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Gagal menghapus produk")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func applyProductPatch(product *model.Product, patch productPatch) {
	if patch.ResetReformatted {
		product.ReformattedText = nil
		product.Status = "raw"
	}
	if patch.SourceCategory != nil {
		product.SourceCategory = *patch.SourceCategory
	}
	if patch.TrackingTag != nil {
		product.TrackingTag = *patch.TrackingTag
	}
	if patch.ShopeeLink != nil {
		product.ShopeeLink = *patch.ShopeeLink
	}
	if patch.ReformattedText != nil {
		product.ReformattedText = patch.ReformattedText
	}
	if patch.ProductName != nil {
		product.ProductName = patch.ProductName
	}
	if patch.ImageURL != nil {
		product.ImageURL = *patch.ImageURL
	}
	if patch.ImageURLs != nil {
		product.ImageURLs = *patch.ImageURLs
	}
	if patch.VideoURL != nil {
		product.VideoURL = *patch.VideoURL
	}
	if patch.NormalPrice != nil {
		product.NormalPrice = *patch.NormalPrice
	}
	if patch.SalePrice != nil {
		product.SalePrice = *patch.SalePrice
	}
	if patch.DiscountPercent != nil {
		product.DiscountPercent = *patch.DiscountPercent
	}
	if patch.Rating != nil {
		product.Rating = *patch.Rating
	}
	if patch.SoldCount != nil {
		product.SoldCount = *patch.SoldCount
	}
	if patch.ReviewCount != nil {
		product.ReviewCount = *patch.ReviewCount
	}
	if patch.Keyword != nil {
		product.Keyword = *patch.Keyword
	}
	if patch.Problem != nil {
		product.Problem = *patch.Problem
	}
	if patch.Cluster != nil {
		product.Cluster = *patch.Cluster
	}
	if patch.ContentModel != nil {
		product.ContentModel = *patch.ContentModel
	}
	if patch.CaptureAngle != nil {
		product.CaptureAngle = *patch.CaptureAngle
	}
	if patch.Benefit1 != nil {
		product.Benefit1 = *patch.Benefit1
	}
	if patch.Benefit2 != nil {
		product.Benefit2 = *patch.Benefit2
	}
	if patch.Benefit3 != nil {
		product.Benefit3 = *patch.Benefit3
	}
	if patch.Urgency != nil {
		product.Urgency = *patch.Urgency
	}
	if patch.CaptionTemplate != nil {
		product.CaptionTemplate = *patch.CaptionTemplate
	}
	if patch.HashtagPool != nil {
		product.HashtagPool = *patch.HashtagPool
	}
	if patch.Notes != nil {
		product.Notes = *patch.Notes
	}
	if patch.Status != nil {
		product.Status = *patch.Status
	}
}

func parsePositive(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
