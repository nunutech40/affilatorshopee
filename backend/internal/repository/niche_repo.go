package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type NicheRepository struct{ db *sql.DB }

func NewNicheRepository(db *sql.DB) *NicheRepository { return &NicheRepository{db: db} }

func (r *NicheRepository) List(ctx context.Context) ([]model.Niche, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at FROM niches ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	niches := []model.Niche{}
	for rows.Next() {
		var niche model.Niche
		if err := rows.Scan(&niche.ID, &niche.Name, &niche.CreatedAt); err != nil {
			return nil, err
		}
		niches = append(niches, niche)
	}
	return niches, rows.Err()
}

func (r *NicheRepository) Create(ctx context.Context, name string) (*model.Niche, error) {
	name = strings.TrimSpace(name)
	var niche model.Niche
	err := r.db.QueryRowContext(ctx, `INSERT INTO niches (name) VALUES ($1) RETURNING id, name, created_at`, name).
		Scan(&niche.ID, &niche.Name, &niche.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &niche, nil
}

func (r *NicheRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM niches WHERE id=$1`, id)
	return err
}

func (r *NicheRepository) ReplaceProductNiches(ctx context.Context, productID string, nicheIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_niches WHERE product_id=$1`, productID); err != nil {
		return err
	}
	for _, nicheID := range nicheIDs {
		if strings.TrimSpace(nicheID) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO product_niches (product_id, niche_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, productID, nicheID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
