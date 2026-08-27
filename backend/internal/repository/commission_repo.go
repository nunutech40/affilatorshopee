package repository

import (
	"context"
	"database/sql"
	"fmt"
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
		err = tx.QueryRowContext(ctx, `INSERT INTO commission_events (event_id, order_id, item_id, model_id, order_status, ordered_at, tracking_tag, normalized_tag, quantity, commission_total, item_name, shop_name)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT (event_id) DO UPDATE SET order_status=EXCLUDED.order_status, ordered_at=EXCLUDED.ordered_at, tracking_tag=EXCLUDED.tracking_tag, normalized_tag=EXCLUDED.normalized_tag, quantity=EXCLUDED.quantity, commission_total=EXCLUDED.commission_total, item_name=EXCLUDED.item_name, shop_name=EXCLUDED.shop_name RETURNING (xmax = 0)`, event.EventID, event.OrderID, event.ItemID, event.ModelID, event.OrderStatus, event.OrderedAt, event.TrackingTag, normalized, event.Quantity, event.CommissionTotal, event.ItemName, event.ShopName).Scan(&created)
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

func (r *CommissionRepository) ListSoldProducts(ctx context.Context, limit, offset int, search string) ([]model.SoldProduct, int, error) {
	where := ""
	args := []interface{}{}
	if s := strings.TrimSpace(search); s != "" {
		where = "WHERE ce.tracking_tag ILIKE $1 OR p.product_name ILIKE $1 OR ce.item_name ILIKE $1"
		args = append(args, "%"+s+"%")
	}
	countQuery := "SELECT COUNT(*) FROM (SELECT ce.normalized_tag FROM commission_events ce LEFT JOIN products p ON regexp_replace(lower(p.tracking_tag), '[^a-z0-9]', '', 'g') = ce.normalized_tag " + where + " GROUP BY ce.normalized_tag) s"
	var total int
	countArgs := args
	if where != "" {
		if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
			return nil, 0, err
		}
	}
	query := `
		SELECT ce.normalized_tag, ce.tracking_tag,
			p.id, p.product_name, p.shopee_link, p.image_url,
			MAX(ce.item_name), MAX(ce.item_id), MAX(ce.shop_name),
			SUM(ce.quantity)::int AS total_quantity,
			SUM(ce.commission_total)::bigint AS total_commission,
			COUNT(*)::int AS order_count,
			MAX(ce.ordered_at) AS last_ordered_at,
			(p.id IS NOT NULL) AS is_in_library
		FROM commission_events ce
		LEFT JOIN products p ON regexp_replace(lower(p.tracking_tag), '[^a-z0-9]', '', 'g') = ce.normalized_tag
		` + where + `
		GROUP BY ce.normalized_tag, ce.tracking_tag, p.id, p.product_name, p.shopee_link, p.image_url
		ORDER BY total_quantity DESC, last_ordered_at DESC
		LIMIT $` + fmtInt(len(args)+1) + ` OFFSET $` + fmtInt(len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []model.SoldProduct{}
	for rows.Next() {
		var s model.SoldProduct
		if err := rows.Scan(&s.NormalizedTag, &s.TrackingTag, &s.ProductID, &s.ProductName, &s.ShopeeLink, &s.ImageURL, &s.ItemName, &s.ItemID, &s.ShopName, &s.TotalQuantity, &s.TotalCommission, &s.OrderCount, &s.LastOrderedAt, &s.IsInLibrary); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func fmtInt(v int) string { return strings.TrimSpace(strings.ReplaceAll(fmt.Sprint(v), " ", "")) }

var _ = strings.TrimSpace
