package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
	"github.com/nunutech40/affilatorshopee/internal/repository"
)

type PostLogService struct {
	repo *repository.PostLogRepository
}

func NewPostLogService(repo *repository.PostLogRepository) *PostLogService {
	return &PostLogService{repo: repo}
}

func (s *PostLogService) Create(ctx context.Context, log *model.PostLog) error {
	if log.ProductID == "" {
		return fmt.Errorf("%w: product_id wajib diisi", ErrValidation)
	}
	log.Caption = strings.TrimSpace(log.Caption)
	if log.Caption == "" {
		return fmt.Errorf("%w: caption wajib diisi", ErrValidation)
	}
	if log.Platform == "" {
		log.Platform = "x"
	}
	if log.Platform != "x" {
		return fmt.Errorf("%w: platform MVP hanya x", ErrValidation)
	}
	if len(log.Hashtags) > 3 {
		return fmt.Errorf("%w: maksimal 3 hashtag", ErrValidation)
	}
	return s.repo.Create(ctx, log)
}

func (s *PostLogService) ListByProduct(ctx context.Context, productID string) ([]model.PostLog, error) {
	return s.repo.ListByProduct(ctx, productID)
}

func (s *PostLogService) ListAll(ctx context.Context) ([]model.PostLog, error) {
	return s.repo.ListAll(ctx)
}
