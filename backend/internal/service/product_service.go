package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, product *model.Product) error {
	if strings.TrimSpace(product.RawText) == "" {
		return fmt.Errorf("%w: raw_text wajib diisi", ErrValidation)
	}
	if strings.TrimSpace(product.ShopeeLink) == "" && product.SourceCategory != "scrape_shopee" {
		return fmt.Errorf("%w: shopee_link wajib diisi", ErrValidation)
	}
	if strings.TrimSpace(product.ShopeeLink) != "" {
		if err := validateURL(product.ShopeeLink); err != nil {
			return err
		}
	}
	if product.SourceCategory == "scrape_shopee" && (product.Notes == nil || strings.TrimSpace(*product.Notes) == "") {
		note := "Sumber: halaman produk Shopee; link affiliate belum diisi."
		product.Notes = &note
	}
	product.Status = "raw"
	if product.SourceCategory == "" {
		product.SourceCategory = "raw_text"
	}
	if product.CaptionTemplate == "" {
		product.CaptionTemplate = "direct_product"
	}
	if strings.TrimSpace(product.TrackingTag) == "" {
		product.TrackingTag = generateTrackingTag(product.RawText)
	}
	if err := validateProductFields(product); err != nil {
		return err
	}
	return s.repo.Create(ctx, product)
}

func generateTrackingTag(raw string) string {
	base := strings.TrimSpace(strings.Split(raw, "\n")[0])
	base = strings.ToLower(base)
	var slug strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			slug.WriteRune(r)
		}
		if slug.Len() >= 24 {
			break
		}
	}
	value := strings.Trim(slug.String(), "-")
	if value == "" {
		value = "produk"
	}
	random := make([]byte, 4)
	if _, err := rand.Read(random); err != nil {
		return value + "tag"
	}
	return value + hex.EncodeToString(random)
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*model.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrNotFound
	}
	return product, nil
}

func (s *ProductService) List(ctx context.Context, filter repository.ProductListFilter) ([]model.Product, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.repo.List(ctx, filter)
}

func (s *ProductService) Update(ctx context.Context, product *model.Product) error {
	if product.ID == "" || !uuidPattern.MatchString(product.ID) {
		return fmt.Errorf("%w: id tidak valid", ErrValidation)
	}
	if strings.TrimSpace(product.RawText) == "" {
		return fmt.Errorf("%w: raw_text tidak boleh kosong", ErrValidation)
	}
	if strings.TrimSpace(product.ShopeeLink) == "" && product.SourceCategory != "scrape_shopee" {
		return fmt.Errorf("%w: shopee_link wajib diisi", ErrValidation)
	}
	if strings.TrimSpace(product.ShopeeLink) != "" {
		if err := validateURL(product.ShopeeLink); err != nil {
			return err
		}
	}
	if err := validateProductFields(product); err != nil {
		return err
	}
	existing, err := s.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}
	if existing.Status != product.Status && !validStatusTransition(existing.Status, product.Status) {
		return fmt.Errorf("%w: transisi status tidak diizinkan", ErrValidation)
	}
	if product.Status == "ready" && existing.Status != "ready" {
		if err := validateReady(product); err != nil {
			return err
		}
	}
	return s.repo.Update(ctx, product)
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	if id == "" || !uuidPattern.MatchString(id) {
		return fmt.Errorf("%w: id tidak valid", ErrValidation)
	}
	return s.repo.Delete(ctx, id)
}

func (s *ProductService) MarkReady(ctx context.Context, id string) error {
	product, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := validateReady(product); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, "ready")
}

func validStatusTransition(from, to string) bool {
	return (from == "raw" && (to == "reformatted" || to == "ready")) ||
		(from == "reformatted" && (to == "ready" || to == "raw")) ||
		(from == "ready" && to == "raw")
}

func validateReady(product *model.Product) error {
	if product.ProductName == nil || strings.TrimSpace(*product.ProductName) == "" {
		return fmt.Errorf("%w: product_name wajib diisi untuk ready", ErrValidation)
	}
	if product.Cluster == nil || strings.TrimSpace(*product.Cluster) == "" {
		return fmt.Errorf("%w: cluster wajib diisi untuk ready", ErrValidation)
	}
	if product.ContentModel == nil || strings.TrimSpace(*product.ContentModel) == "" {
		return fmt.Errorf("%w: content_model wajib diisi untuk ready", ErrValidation)
	}
	normalized := normalizeContentModel(*product.ContentModel)
	if !hasText(product.Benefit1) && !hasText(product.Benefit2) && !hasText(product.Benefit3) {
		return fmt.Errorf("%w: minimal satu benefit wajib diisi untuk ready", ErrValidation)
	}
	if normalized == "capture" && (product.CaptureAngle == nil || strings.TrimSpace(*product.CaptureAngle) == "") {
		return fmt.Errorf("%w: capture_angle wajib diisi untuk content_model capture", ErrValidation)
	}
	return nil
}

func hasText(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != ""
}

func normalizeContentModel(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "captured" {
		return "capture"
	}
	if normalized == "murah" || normalized == "value" {
		return "cheap"
	}
	return normalized
}

func validateProductFields(product *model.Product) error {
	if product.ContentModel != nil {
		normalized := normalizeContentModel(*product.ContentModel)
		*product.ContentModel = normalized
	}
	if product.Status != "raw" && product.Status != "reformatted" && product.Status != "ready" {
		return fmt.Errorf("%w: status tidak valid", ErrValidation)
	}
	if product.CaptionTemplate != "direct_product" && product.CaptionTemplate != "keyword_recommendation" && product.CaptionTemplate != "problem_specific" && product.CaptionTemplate != "cheap_value" {
		return fmt.Errorf("%w: caption_template tidak valid", ErrValidation)
	}
	if product.ContentModel != nil && *product.ContentModel != "capture" && *product.ContentModel != "cheap" && *product.ContentModel != "trending" && *product.ContentModel != "branded" {
		return fmt.Errorf("%w: content_model tidak valid", ErrValidation)
	}
	if product.CaptureAngle != nil {
		valid := map[string]bool{"search": true, "reply": true, "trend": true, "problem": true}
		if !valid[*product.CaptureAngle] || product.ContentModel == nil || normalizeContentModel(*product.ContentModel) != "capture" {
			return fmt.Errorf("%w: capture_angle tidak valid", ErrValidation)
		}
	}
	if product.NormalPrice != nil && *product.NormalPrice < 0 || product.SalePrice != nil && *product.SalePrice < 0 {
		return fmt.Errorf("%w: harga tidak valid", ErrValidation)
	}
	if product.NormalPrice != nil && product.SalePrice != nil && *product.SalePrice > *product.NormalPrice {
		return fmt.Errorf("%w: sale_price lebih besar dari normal_price", ErrValidation)
	}
	if product.DiscountPercent != nil && (*product.DiscountPercent < 0 || *product.DiscountPercent > 100) {
		return fmt.Errorf("%w: discount_percent tidak valid", ErrValidation)
	}
	if product.Rating != nil && (*product.Rating < 0 || *product.Rating > 5) {
		return fmt.Errorf("%w: rating tidak valid", ErrValidation)
	}
	if len(product.HashtagPool) > 3 {
		return fmt.Errorf("%w: hashtag_pool maksimal 3", ErrValidation)
	}
	return nil
}

func validateURL(value string) error {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%w: URL harus menggunakan http atau https", ErrValidation)
	}
	return nil
}

type ReformatFailure struct {
	ProductID string `json:"product_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type ReformatSummary struct {
	Processed []model.Product   `json:"processed"`
	Failed    []ReformatFailure `json:"failed"`
}

func (s *ProductService) Reformat(ctx context.Context, ids []string, ai *AIService, modelOverride string, variant ...bool) (ReformatSummary, error) {
	if len(ids) < 1 || len(ids) > 10 {
		return ReformatSummary{}, fmt.Errorf("%w: product_ids harus berisi 1-10 ID (maksimal 10 untuk hemat token)", ErrValidation)
	}
	seen := make(map[string]struct{}, len(ids))
	products := make([]model.Product, 0, len(ids))
	summary := ReformatSummary{}
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			return ReformatSummary{}, fmt.Errorf("%w: product_id duplikat", ErrValidation)
		}
		seen[id] = struct{}{}
		product, err := s.GetByID(ctx, id)
		if err == ErrNotFound {
			summary.Failed = append(summary.Failed, ReformatFailure{ProductID: id, Code: "PRODUCT_NOT_FOUND", Message: "Produk tidak ditemukan"})
			continue
		}
		if err != nil {
			return ReformatSummary{}, err
		}
		if product.Status != "raw" && product.Status != "reformatted" && product.Status != "ready" {
			summary.Failed = append(summary.Failed, ReformatFailure{ProductID: id, Code: "PRODUCT_STATUS_INVALID", Message: "Status produk tidak bisa direformat"})
			continue
		}
		products = append(products, *product)
	}
	if len(products) == 0 {
		return summary, nil
	}

	results, err := ai.Reformat(ctx, products, modelOverride, variant...)
	if err != nil {
		return ReformatSummary{}, err
	}
	byID := make(map[string]AIReformatResult, len(results))
	for _, result := range results {
		byID[result.ProductID] = result
	}
	for _, product := range products {
		result, ok := byID[product.ID]
		if !ok {
			summary.Failed = append(summary.Failed, ReformatFailure{ProductID: product.ID, Code: "AI_MISSING_RESULT", Message: "AI tidak mengembalikan hasil untuk produk"})
			continue
		}
		applyAIResult(&product, result)
		if product.Status == "raw" && strings.TrimSpace(result.PromoText) != "" {
			product.Status = "reformatted"
		}
		if err := s.repo.Update(ctx, &product); err != nil {
			if err == sql.ErrNoRows {
				summary.Failed = append(summary.Failed, ReformatFailure{ProductID: product.ID, Code: "PRODUCT_CHANGED", Message: "Produk berubah sebelum hasil AI disimpan"})
				continue
			}
			return ReformatSummary{}, err
		}
		summary.Processed = append(summary.Processed, product)
	}
	return summary, nil
}

func applyAIResult(product *model.Product, result AIReformatResult) {
	if strings.TrimSpace(result.PromoText) != "" {
		txt := strings.TrimSpace(result.PromoText)
		product.ReformattedText = &txt
	}
	// legacy compat: jika AI masih kirim field terstruktur, tetap simpan
	if result.ProductName != nil {
		product.ProductName = result.ProductName
	}
	if result.NormalPrice != nil {
		product.NormalPrice = result.NormalPrice
	}
	if result.SalePrice != nil {
		product.SalePrice = result.SalePrice
	}
	if result.DiscountPercent != nil {
		product.DiscountPercent = result.DiscountPercent
	}
	if result.Rating != nil {
		product.Rating = result.Rating
	}
	if result.SoldCount != nil {
		product.SoldCount = result.SoldCount
	}
	if result.ReviewCount != nil {
		product.ReviewCount = result.ReviewCount
	}
	if result.Keyword != nil {
		product.Keyword = result.Keyword
	}
	if result.Problem != nil {
		product.Problem = result.Problem
	}
	if result.Cluster != nil {
		product.Cluster = result.Cluster
	}
	if result.ContentModel != nil {
		product.ContentModel = result.ContentModel
	}
	if result.CaptureAngle != nil {
		product.CaptureAngle = result.CaptureAngle
	}
	if result.Benefit1 != nil {
		product.Benefit1 = result.Benefit1
	}
	if result.Benefit2 != nil {
		product.Benefit2 = result.Benefit2
	}
	if result.Benefit3 != nil {
		product.Benefit3 = result.Benefit3
	}
	if result.Urgency != nil {
		product.Urgency = result.Urgency
	}
	if len(result.HashtagPool) > 0 {
		product.HashtagPool = result.HashtagPool
	}
}

func RuneLen(value string) int {
	return utf8.RuneCountInString(value)
}
