package service

import (
	"context"
	"database/sql"
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
	if strings.TrimSpace(product.ShopeeLink) == "" {
		return fmt.Errorf("%w: shopee_link wajib diisi", ErrValidation)
	}
	if err := validateURL(product.ShopeeLink); err != nil {
		return err
	}
	if err := validateProductFields(product); err != nil {
		return err
	}
	product.Status = "raw"
	if product.CaptionTemplate == "" {
		product.CaptionTemplate = "direct_product"
	}
	return s.repo.Create(ctx, product)
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
	if strings.TrimSpace(product.ShopeeLink) == "" {
		return fmt.Errorf("%w: shopee_link wajib diisi", ErrValidation)
	}
	if err := validateURL(product.ShopeeLink); err != nil {
		return err
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
	if product.Status == "ready" {
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
	return (from == "raw" && (to == "reformatted" || to == "ready")) || (from == "reformatted" && to == "ready")
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
	if !hasText(product.Benefit1) && !hasText(product.Benefit2) && !hasText(product.Benefit3) {
		return fmt.Errorf("%w: minimal satu benefit wajib diisi untuk ready", ErrValidation)
	}
	if *product.ContentModel == "capture" && (product.CaptureAngle == nil || strings.TrimSpace(*product.CaptureAngle) == "") {
		return fmt.Errorf("%w: capture_angle wajib diisi untuk content_model capture", ErrValidation)
	}
	return nil
}

func hasText(value *string) bool {
	return value != nil && strings.TrimSpace(*value) != ""
}

func validateProductFields(product *model.Product) error {
	if product.Status != "raw" && product.Status != "reformatted" && product.Status != "ready" {
		return fmt.Errorf("%w: status tidak valid", ErrValidation)
	}
	if product.CaptionTemplate != "direct_product" && product.CaptionTemplate != "keyword_recommendation" && product.CaptionTemplate != "problem_specific" && product.CaptionTemplate != "cheap_value" {
		return fmt.Errorf("%w: caption_template tidak valid", ErrValidation)
	}
	if product.ContentModel != nil && *product.ContentModel != "capture" && *product.ContentModel != "cheap" {
		return fmt.Errorf("%w: content_model tidak valid", ErrValidation)
	}
	if product.CaptureAngle != nil {
		valid := map[string]bool{"search": true, "reply": true, "trend": true, "problem": true}
		if !valid[*product.CaptureAngle] || product.ContentModel == nil || *product.ContentModel != "capture" {
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

func (s *ProductService) Reformat(ctx context.Context, ids []string, ai *AIService) (ReformatSummary, error) {
	if len(ids) < 1 || len(ids) > 20 {
		return ReformatSummary{}, fmt.Errorf("%w: product_ids harus berisi 1-20 ID", ErrValidation)
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
		if product.Status != "raw" {
			summary.Failed = append(summary.Failed, ReformatFailure{ProductID: id, Code: "PRODUCT_NOT_RAW", Message: "Produk bukan berstatus raw"})
			continue
		}
		products = append(products, *product)
	}
	if len(products) == 0 {
		return summary, nil
	}

	results, err := ai.Reformat(ctx, products)
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
		product.Status = "reformatted"
		if err := s.repo.UpdateReformatted(ctx, &product); err != nil {
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
	product.ProductName = result.ProductName
	product.NormalPrice = result.NormalPrice
	product.SalePrice = result.SalePrice
	product.DiscountPercent = result.DiscountPercent
	product.Rating = result.Rating
	product.SoldCount = result.SoldCount
	product.ReviewCount = result.ReviewCount
	product.Keyword = result.Keyword
	product.Problem = result.Problem
	product.Cluster = result.Cluster
	product.ContentModel = result.ContentModel
	product.CaptureAngle = result.CaptureAngle
	product.Benefit1 = result.Benefit1
	product.Benefit2 = result.Benefit2
	product.Benefit3 = result.Benefit3
	product.Urgency = result.Urgency
	product.HashtagPool = result.HashtagPool
}

func RuneLen(value string) int {
	return utf8.RuneCountInString(value)
}
