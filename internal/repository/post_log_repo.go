package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/nunutech40/affilatorshopee/internal/model"
)

type PostLogRepository struct {
	db *sql.DB
}

func NewPostLogRepository(db *sql.DB) *PostLogRepository {
	return &PostLogRepository{db: db}
}

func (r *PostLogRepository) Create(ctx context.Context, log *model.PostLog) error {
	query := `INSERT INTO post_logs (product_id, platform, caption, hashtags, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, posted_at`
	return r.db.QueryRowContext(ctx, query, log.ProductID, log.Platform, log.Caption, pq.Array(log.Hashtags), log.Notes).
		Scan(&log.ID, &log.PostedAt)
}

func (r *PostLogRepository) ListByProduct(ctx context.Context, productID string) ([]model.PostLog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, platform, caption, hashtags, notes, posted_at
		FROM post_logs WHERE product_id=$1 ORDER BY posted_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []model.PostLog{}
	for rows.Next() {
		var log model.PostLog
		if err := rows.Scan(&log.ID, &log.ProductID, &log.Platform, &log.Caption, pq.Array(&log.Hashtags), &log.Notes, &log.PostedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (r *PostLogRepository) ListAll(ctx context.Context) ([]model.PostLog, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, product_id, platform, caption, hashtags, notes, posted_at
		FROM post_logs ORDER BY posted_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := []model.PostLog{}
	for rows.Next() {
		var log model.PostLog
		if err := rows.Scan(&log.ID, &log.ProductID, &log.Platform, &log.Caption, pq.Array(&log.Hashtags), &log.Notes, &log.PostedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
