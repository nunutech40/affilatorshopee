package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/nunutech40/affilatorshopee/internal/model"
)

type CaptionVariationRepository struct {
	db *sql.DB
}

func NewCaptionVariationRepository(db *sql.DB) *CaptionVariationRepository {
	return &CaptionVariationRepository{db: db}
}

func (r *CaptionVariationRepository) Create(ctx context.Context, variation *model.CaptionVariation) error {
	query := `INSERT INTO caption_variations (product_id, label, template, caption, hashtags)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, variation.ProductID, variation.Label, variation.Template, variation.Caption, pq.Array(variation.Hashtags)).
		Scan(&variation.ID, &variation.CreatedAt, &variation.UpdatedAt)
}

func (r *CaptionVariationRepository) ListByProduct(ctx context.Context, productID string) ([]model.CaptionVariation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, label, template, caption, hashtags, created_at, updated_at
		FROM caption_variations WHERE product_id=$1 ORDER BY created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	variations := []model.CaptionVariation{}
	for rows.Next() {
		var variation model.CaptionVariation
		if err := rows.Scan(&variation.ID, &variation.ProductID, &variation.Label, &variation.Template, &variation.Caption, pq.Array(&variation.Hashtags), &variation.CreatedAt, &variation.UpdatedAt); err != nil {
			return nil, err
		}
		variations = append(variations, variation)
	}
	return variations, rows.Err()
}

func (r *CaptionVariationRepository) GetByID(ctx context.Context, id string) (*model.CaptionVariation, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, product_id, label, template, caption, hashtags, created_at, updated_at
		FROM caption_variations WHERE id=$1`, id)
	var variation model.CaptionVariation
	if err := row.Scan(&variation.ID, &variation.ProductID, &variation.Label, &variation.Template, &variation.Caption, pq.Array(&variation.Hashtags), &variation.CreatedAt, &variation.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &variation, nil
}

func (r *CaptionVariationRepository) Update(ctx context.Context, variation *model.CaptionVariation) error {
	_, err := r.db.ExecContext(ctx, `UPDATE caption_variations SET label=$2, caption=$3, hashtags=$4 WHERE id=$1`, variation.ID, variation.Label, variation.Caption, pq.Array(variation.Hashtags))
	return err
}

func (r *CaptionVariationRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM caption_variations WHERE id=$1`, id)
	return err
}
