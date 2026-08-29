package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nunutech40/affilatorshopee/internal/model"
)

type ContentRepository struct{ db *sql.DB }

func NewContentRepository(db *sql.DB) *ContentRepository { return &ContentRepository{db: db} }

func contentSlug(name string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(name)))
	return strings.Join(parts, "-")
}

func (r *ContentRepository) ListNiches(ctx context.Context) ([]model.ContentNiche, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, description, status, created_at, updated_at FROM content_niches WHERE status='active' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []model.ContentNiche{}
	for rows.Next() {
		var item model.ContentNiche
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ContentRepository) CreateNiche(ctx context.Context, name string) (*model.ContentNiche, error) {
	var item model.ContentNiche
	err := r.db.QueryRowContext(ctx, `INSERT INTO content_niches (name, slug) VALUES ($1,$2) RETURNING id,name,slug,description,status,created_at,updated_at`, strings.TrimSpace(name), contentSlug(name)).Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	return &item, err
}

func (r *ContentRepository) DeleteNiche(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE content_niches SET status='archived', updated_at=CURRENT_TIMESTAMP WHERE id=$1`, id)
	return err
}

func (r *ContentRepository) Create(ctx context.Context, item model.ContentItem, nicheIDs, productTypeIDs []string, stats *model.ContentStats) (*model.ContentItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var created model.ContentItem
	mediaBytes, err := json.Marshal(item.Media)
	if err != nil {
		return nil, err
	}
	media := string(mediaBytes)
	err = tx.QueryRowContext(ctx, `INSERT INTO content_items (platform,external_post_id,canonical_url,author_handle,original_text,media,published_at,source_query,status) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9) RETURNING id,platform,external_post_id,canonical_url,author_handle,original_text,created_at,updated_at`, item.Platform, item.ExternalPostID, item.CanonicalURL, item.AuthorHandle, item.OriginalText, media, item.PublishedAt, item.SourceQuery, item.Status).Scan(&created.ID, &created.Platform, &created.ExternalPostID, &created.CanonicalURL, &created.AuthorHandle, &created.OriginalText, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	for _, id := range nicheIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_item_niches (content_item_id,content_niche_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, created.ID, id); err != nil {
			return nil, err
		}
	}
	for _, id := range productTypeIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_item_product_types (content_item_id,product_type_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, created.ID, id); err != nil {
			return nil, err
		}
	}
	if stats != nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_stat_snapshots (content_item_id,like_count,repost_count,reply_count,bookmark_count,view_count) VALUES ($1,$2,$3,$4,$5,$6)`, created.ID, stats.LikeCount, stats.RepostCount, stats.ReplyCount, stats.BookmarkCount, stats.ViewCount); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, created.ID)
}

func (r *ContentRepository) List(ctx context.Context, nicheID, platform, status, search string, page, limit int) ([]model.ContentItem, int, error) {
	args := []interface{}{}
	where := []string{"1=1"}
	n := 1
	if nicheID != "" {
		where = append(where, fmt.Sprintf("EXISTS (SELECT 1 FROM content_item_niches cin WHERE cin.content_item_id=ci.id AND cin.content_niche_id=$%d)", n))
		args = append(args, nicheID)
		n++
	}
	if platform != "" {
		where = append(where, fmt.Sprintf("ci.platform=$%d", n))
		args = append(args, platform)
		n++
	}
	if status != "" {
		where = append(where, fmt.Sprintf("ci.status=$%d", n))
		args = append(args, status)
		n++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(ci.original_text ILIKE $%d OR ci.canonical_url ILIKE $%d)", n, n))
		args = append(args, "%"+search+"%")
		n++
	}
	countArgs := append([]interface{}{}, args...)
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM content_items ci WHERE `+strings.Join(where, " AND "), countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, (page-1)*limit)
	rows, err := r.db.QueryContext(ctx, `SELECT ci.id,ci.platform,ci.external_post_id,ci.canonical_url,ci.author_handle,ci.original_text,ci.published_at,ci.source_query,ci.status,ci.created_at,ci.updated_at FROM content_items ci WHERE `+strings.Join(where, " AND ")+fmt.Sprintf(" ORDER BY ci.created_at DESC LIMIT $%d OFFSET $%d", n, n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []model.ContentItem{}
	for rows.Next() {
		var i model.ContentItem
		if err := rows.Scan(&i.ID, &i.Platform, &i.ExternalPostID, &i.CanonicalURL, &i.AuthorHandle, &i.OriginalText, &i.PublishedAt, &i.SourceQuery, &i.Status, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, rows.Err()
}

func (r *ContentRepository) Get(ctx context.Context, id string) (*model.ContentItem, error) {
	var i model.ContentItem
	err := r.db.QueryRowContext(ctx, `SELECT id,platform,external_post_id,canonical_url,author_handle,original_text,published_at,source_query,status,created_at,updated_at FROM content_items WHERE id=$1`, id).Scan(&i.ID, &i.Platform, &i.ExternalPostID, &i.CanonicalURL, &i.AuthorHandle, &i.OriginalText, &i.PublishedAt, &i.SourceQuery, &i.Status, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &i, nil
}
