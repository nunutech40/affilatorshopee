package repository

import (
	"context"
	"database/sql"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type MediaRepository struct{ db *sql.DB }

func NewMediaRepository(db *sql.DB) *MediaRepository { return &MediaRepository{db: db} }

func (r *MediaRepository) Create(ctx context.Context, media *model.MediaFile) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO product_media (product_id, source_url, local_path, filename, media_type, content_type, size_bytes)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`, media.ProductID, media.SourceURL, media.LocalPath, media.Filename, media.MediaType, media.ContentType, media.SizeBytes).
		Scan(&media.ID, &media.CreatedAt)
}

func (r *MediaRepository) ListByProduct(ctx context.Context, productID string) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, source_url, local_path, filename, media_type, content_type, size_bytes, created_at
		FROM product_media WHERE product_id=$1 ORDER BY created_at ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.MediaFile{}
	for rows.Next() {
		var media model.MediaFile
		if err := rows.Scan(&media.ID, &media.ProductID, &media.SourceURL, &media.LocalPath, &media.Filename, &media.MediaType, &media.ContentType, &media.SizeBytes, &media.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, media)
	}
	return items, rows.Err()
}

func (r *MediaRepository) GetByID(ctx context.Context, productID, mediaID string) (*model.MediaFile, error) {
	var media model.MediaFile
	err := r.db.QueryRowContext(ctx, `SELECT id, product_id, source_url, local_path, filename, media_type, content_type, size_bytes, created_at
		FROM product_media WHERE id=$1 AND product_id=$2`, mediaID, productID).
		Scan(&media.ID, &media.ProductID, &media.SourceURL, &media.LocalPath, &media.Filename, &media.MediaType, &media.ContentType, &media.SizeBytes, &media.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *MediaRepository) Delete(ctx context.Context, productID, mediaID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product_media WHERE id=$1 AND product_id=$2`, mediaID, productID)
	return err
}
