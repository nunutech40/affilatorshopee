package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type ClickSyncResult struct {
	Imported  int      `json:"imported"`
	Duplicate int      `json:"duplicate"`
	Matched   int      `json:"matched"`
	Unmatched []string `json:"unmatched_tags"`
}

type ClickRepository struct{ db *sql.DB }

func NewClickRepository(db *sql.DB) *ClickRepository { return &ClickRepository{db: db} }

func normalizeTrackingTag(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (r *ClickRepository) Sync(ctx context.Context, events []model.ClickEvent) (ClickSyncResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ClickSyncResult{}, err
	}
	defer tx.Rollback()
	result := ClickSyncResult{}
	seenUnmatched := map[string]bool{}
	for _, event := range events {
		normalized := normalizeTrackingTag(event.TrackingTag)
		if normalized == "" || event.ClickID == "" {
			continue
		}
		var inserted bool
		err = tx.QueryRowContext(ctx, `INSERT INTO click_events (click_id, clicked_at, region, tracking_tag, normalized_tag, referrer)
			VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (click_id) DO NOTHING RETURNING true`, event.ClickID, event.ClickedAt, event.Region, event.TrackingTag, normalized, event.Referrer).Scan(&inserted)
		if err == sql.ErrNoRows {
			result.Duplicate++
			continue
		}
		if err != nil {
			return result, err
		}
		result.Imported++
		var matched bool
		err = tx.QueryRowContext(ctx, `UPDATE products SET click_count=click_count+1, last_clicked_at=$2 WHERE regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g')=$1 RETURNING true`, normalized, event.ClickedAt).Scan(&matched)
		if err == sql.ErrNoRows {
			var archived bool
			if archiveErr := tx.QueryRowContext(ctx, `UPDATE product_tracking_archive SET click_count=click_count+1, last_clicked_at=$2 WHERE regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g')=$1 RETURNING true`, normalized, event.ClickedAt).Scan(&archived); archiveErr == nil && archived {
				result.Matched++
				continue
			}
			if !seenUnmatched[normalized] {
				result.Unmatched = append(result.Unmatched, event.TrackingTag)
				seenUnmatched[normalized] = true
			}
			continue
		}
		if err != nil {
			return result, err
		}
		result.Matched++
	}
	if err := tx.Commit(); err != nil {
		return result, err
	}
	return result, nil
}
