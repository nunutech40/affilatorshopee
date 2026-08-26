package service

import (
	"context"
	"fmt"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type CaptionVariationService struct {
	repo *repository.CaptionVariationRepository
}

func NewCaptionVariationService(repo *repository.CaptionVariationRepository) *CaptionVariationService {
	return &CaptionVariationService{repo: repo}
}

func (s *CaptionVariationService) Create(ctx context.Context, variation *model.CaptionVariation) error {
	if variation.ProductID == "" {
		return fmt.Errorf("%w: product_id wajib diisi", ErrValidation)
	}
	if variation.Caption == "" {
		return fmt.Errorf("%w: caption wajib diisi", ErrValidation)
	}
	if len(variation.Hashtags) > 3 {
		return fmt.Errorf("%w: maksimal 3 hashtag", ErrValidation)
	}
	return s.repo.Create(ctx, variation)
}

func (s *CaptionVariationService) ListByProduct(ctx context.Context, productID string) ([]model.CaptionVariation, error) {
	return s.repo.ListByProduct(ctx, productID)
}

func (s *CaptionVariationService) GetByID(ctx context.Context, id string) (*model.CaptionVariation, error) {
	variation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if variation == nil {
		return nil, ErrNotFound
	}
	return variation, nil
}

func (s *CaptionVariationService) Update(ctx context.Context, variation *model.CaptionVariation) error {
	if variation.ID == "" {
		return fmt.Errorf("%w: id wajib diisi", ErrValidation)
	}
	if variation.Caption == "" {
		return fmt.Errorf("%w: caption wajib diisi", ErrValidation)
	}
	if len(variation.Hashtags) > 3 {
		return fmt.Errorf("%w: maksimal 3 hashtag", ErrValidation)
	}
	return s.repo.Update(ctx, variation)
}

func (s *CaptionVariationService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("%w: id wajib diisi", ErrValidation)
	}
	return s.repo.Delete(ctx, id)
}
