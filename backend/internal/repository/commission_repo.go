package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type CommissionSyncResult struct {
	Imported  int      `json:"imported"`
	Updated   int      `json:"updated"`
	Matched   int      `json:"matched"`
	Unmatched []string `json:"unmatched_tags"`
}

type CommissionRepository struct{ db *sql.DB }

func NewCommissionRepository(db *sql.DB) *CommissionRepository { return &CommissionRepository{db: db} }

func (r *CommissionRepository) Sync(ctx context.Context, events []model.CommissionEvent) (CommissionSyncResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CommissionSyncResult{}, err
	}
	defer tx.Rollback()
	result := CommissionSyncResult{}
	seen := map[string]bool{}
	for _, event := range events {
		normalized := normalizeTrackingTag(event.TrackingTag)
		if event.EventID == "" || normalized == "" {
			continue
		}
		var created bool
		err = tx.QueryRowContext(ctx, `INSERT INTO commission_events (event_id, order_id, item_id, model_id, order_status, ordered_at, tracking_tag, normalized_tag, quantity, commission_total)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (event_id) DO UPDATE SET order_status=EXCLUDED.order_status, ordered_at=EXCLUDED.ordered_at, tracking_tag=EXCLUDED.tracking_tag, normalized_tag=EXCLUDED.normalized_tag, quantity=EXCLUDED.quantity, commission_total=EXCLUDED.commission_total RETURNING (xmax = 0)`, event.EventID, event.OrderID, event.ItemID, event.ModelID, event.OrderStatus, event.OrderedAt, event.TrackingTag, normalized, event.Quantity, event.CommissionTotal).Scan(&created)
		if err != nil {
			return result, err
		}
		if created {
			result.Imported++
		} else {
			result.Updated++
		}
		if !seen[normalized] {
			var exists bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM products WHERE regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g')=$1)`, normalized).Scan(&exists); err != nil {
				return result, err
			}
			if exists {
				result.Matched++
			} else {
				result.Unmatched = append(result.Unmatched, event.TrackingTag)
			}
			seen[normalized] = true
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE products p SET
		sales_count=COALESCE(x.sales_count,0), pending_sales_count=COALESCE(x.pending_sales_count,0), commission_total=COALESCE(x.commission_total,0)
		FROM (SELECT regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g') AS tag,
			SUM(CASE WHEN lower(order_status) IN ('selesai','completed') THEN quantity ELSE 0 END)::int AS sales_count,
			SUM(CASE WHEN lower(order_status) IN ('tertunda','pending') THEN quantity ELSE 0 END)::int AS pending_sales_count,
			SUM(commission_total)::bigint AS commission_total
			FROM commission_events GROUP BY regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g')) x
		WHERE regexp_replace(lower(p.tracking_tag), '[^a-z0-9]', '', 'g')=x.tag`)
	if err != nil {
		return result, err
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}

var _ = strings.TrimSpace
