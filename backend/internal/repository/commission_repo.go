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
		if event.EventID == "" {
			continue
		}
		// The first importer used order_id as the event ID. Newer imports use
		// order_id|item_id|model_id so multiple items in one order remain
		// distinct. Reconcile the legacy row before upserting; otherwise the
		// same commission is counted twice when an old CSV is imported again.
		if event.OrderID != "" && event.ItemID != "" && event.ModelID != "" {
			canonicalID := event.OrderID + "|" + event.ItemID + "|" + event.ModelID
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM commission_events legacy
				WHERE legacy.order_id=$1 AND legacy.item_id=$2 AND legacy.model_id=$3
				  AND legacy.event_id<>$4
				  AND EXISTS (SELECT 1 FROM commission_events canonical WHERE canonical.event_id=$4)`,
				event.OrderID, event.ItemID, event.ModelID, canonicalID); err != nil {
				return result, err
			}
			if _, err := tx.ExecContext(ctx, `
				UPDATE commission_events
				SET event_id=$4
				WHERE order_id=$1 AND item_id=$2 AND model_id=$3
				  AND event_id<>$4
				  AND NOT EXISTS (SELECT 1 FROM commission_events WHERE event_id=$4)`,
				event.OrderID, event.ItemID, event.ModelID, canonicalID); err != nil {
				return result, err
			}
			event.EventID = canonicalID
		}
		if normalized == "" {
			// fallback for Shopee orders without tag - use item_id or order_id to keep unique
			if event.ItemID != "" {
				normalized = "item_" + normalizeTrackingTag(event.ItemID)
			} else {
				normalized = "order_" + normalizeTrackingTag(event.EventID)
			}
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

func (r *CommissionRepository) ListSoldProducts(ctx context.Context, limit, offset int, search string, start, end *string) ([]model.SoldProduct, int, error) {
	whereParts := []string{}
	args := []interface{}{}
	argIdx := 1
	if s := strings.TrimSpace(search); s != "" {
		whereParts = append(whereParts, "(ce.tracking_tag ILIKE $"+fmtInt(argIdx)+" OR p.product_name ILIKE $"+fmtInt(argIdx)+" OR ce.item_name ILIKE $"+fmtInt(argIdx)+")")
		args = append(args, "%"+s+"%")
		argIdx++
	}
	if start != nil && *start != "" {
		whereParts = append(whereParts, "ce.ordered_at >= $"+fmtInt(argIdx))
		args = append(args, *start)
		argIdx++
	}
	if end != nil && *end != "" {
		whereParts = append(whereParts, "ce.ordered_at <= $"+fmtInt(argIdx))
		args = append(args, *end)
		argIdx++
	}
	where := ""
	if len(whereParts) > 0 {
		where = "WHERE " + strings.Join(whereParts, " AND ")
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

func (r *CommissionRepository) ListEvents(ctx context.Context, limit, offset int, search string, start, end *string) ([]model.CommissionEvent, int, error) {
	whereParts := []string{}
	args := []interface{}{}
	argIdx := 1
	if s := strings.TrimSpace(search); s != "" {
		whereParts = append(whereParts, "(item_name ILIKE $"+fmtInt(argIdx)+" OR tracking_tag ILIKE $"+fmtInt(argIdx)+" OR shop_name ILIKE $"+fmtInt(argIdx)+" OR order_id ILIKE $"+fmtInt(argIdx)+")")
		args = append(args, "%"+s+"%")
		argIdx++
	}
	if start != nil && *start != "" {
		whereParts = append(whereParts, "ordered_at >= $"+fmtInt(argIdx))
		args = append(args, *start)
		argIdx++
	}
	if end != nil && *end != "" {
		whereParts = append(whereParts, "ordered_at <= $"+fmtInt(argIdx))
		args = append(args, *end)
		argIdx++
	}
	where := ""
	if len(whereParts) > 0 {
		where = "WHERE " + strings.Join(whereParts, " AND ")
	}
	var total int
	if where != "" {
		if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commission_events "+where, args...).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commission_events").Scan(&total); err != nil {
			return nil, 0, err
		}
	}
	query := "SELECT event_id, order_id, item_id, model_id, order_status, ordered_at, tracking_tag, normalized_tag, quantity, commission_total, item_name, shop_name, imported_at FROM commission_events " + where + " ORDER BY ordered_at DESC NULLS LAST, imported_at DESC LIMIT $" + fmtInt(len(args)+1) + " OFFSET $" + fmtInt(len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []model.CommissionEvent{}
	for rows.Next() {
		var e model.CommissionEvent
		var orderedAt sql.NullTime
		var importedAt sql.NullTime
		if err := rows.Scan(&e.EventID, &e.OrderID, &e.ItemID, &e.ModelID, &e.OrderStatus, &orderedAt, &e.TrackingTag, &e.NormalizedTag, &e.Quantity, &e.CommissionTotal, &e.ItemName, &e.ShopName, &importedAt); err != nil {
			return nil, 0, err
		}
		if orderedAt.Valid {
			e.OrderedAt = &orderedAt.Time
		}
		items = append(items, e)
	}
	return items, total, rows.Err()
}

type CommissionSummary struct {
	TotalCommission int64 `json:"total_commission"`
	TotalQuantity   int   `json:"total_quantity"`
	TotalOrders     int   `json:"total_orders"`
	TotalProducts   int   `json:"total_products"`
}

func (r *CommissionRepository) GetSummary(ctx context.Context, search string, start, end *string) (CommissionSummary, error) {
	whereParts := []string{}
	args := []interface{}{}
	argIdx := 1
	if s := strings.TrimSpace(search); s != "" {
		whereParts = append(whereParts, "(item_name ILIKE $"+fmtInt(argIdx)+" OR tracking_tag ILIKE $"+fmtInt(argIdx)+" OR shop_name ILIKE $"+fmtInt(argIdx)+")")
		args = append(args, "%"+s+"%")
		argIdx++
	}
	if start != nil && *start != "" {
		whereParts = append(whereParts, "ordered_at >= $"+fmtInt(argIdx))
		args = append(args, *start)
		argIdx++
	}
	if end != nil && *end != "" {
		whereParts = append(whereParts, "ordered_at <= $"+fmtInt(argIdx))
		args = append(args, *end)
		argIdx++
	}
	where := ""
	if len(whereParts) > 0 {
		where = "WHERE " + strings.Join(whereParts, " AND ")
	}
	var s CommissionSummary
	// total orders, quantity, commission
	if err := r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(quantity),0)::int, COALESCE(SUM(commission_total),0)::bigint, COUNT(*)::int FROM commission_events "+where, args...).Scan(&s.TotalQuantity, &s.TotalCommission, &s.TotalOrders); err != nil {
		return s, err
	}
	// distinct products (by normalized_tag)
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT normalized_tag) FROM commission_events "+where, args...).Scan(&s.TotalProducts); err != nil {
		return s, err
	}
	return s, nil
}

func fmtInt(v int) string { return strings.TrimSpace(strings.ReplaceAll(fmt.Sprint(v), " ", "")) }

var _ = strings.TrimSpace
