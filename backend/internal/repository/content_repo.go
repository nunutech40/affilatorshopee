package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	err = tx.QueryRowContext(ctx, `INSERT INTO content_items (platform,content_format,external_post_id,canonical_url,author_handle,original_text,media,published_at,source_query,status) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9,$10) RETURNING id,platform,content_format,external_post_id,canonical_url,author_handle,original_text,created_at,updated_at`, item.Platform, item.ContentFormat, item.ExternalPostID, item.CanonicalURL, item.AuthorHandle, item.OriginalText, media, item.PublishedAt, item.SourceQuery, item.Status).Scan(&created.ID, &created.Platform, &created.ContentFormat, &created.ExternalPostID, &created.CanonicalURL, &created.AuthorHandle, &created.OriginalText, &created.CreatedAt, &created.UpdatedAt)
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
	rows, err := r.db.QueryContext(ctx, `SELECT ci.id,ci.platform,ci.content_format,ci.external_post_id,ci.canonical_url,ci.author_handle,ci.original_text,ci.cleaned_original_text,ci.published_at,ci.source_query,ci.status,ci.created_at,ci.updated_at FROM content_items ci WHERE `+strings.Join(where, " AND ")+fmt.Sprintf(" ORDER BY ci.created_at DESC LIMIT $%d OFFSET $%d", n, n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []model.ContentItem{}
	for rows.Next() {
		var i model.ContentItem
		if err := rows.Scan(&i.ID, &i.Platform, &i.ContentFormat, &i.ExternalPostID, &i.CanonicalURL, &i.AuthorHandle, &i.OriginalText, &i.CleanedOriginalText, &i.PublishedAt, &i.SourceQuery, &i.Status, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	for idx := range items {
		if err := r.hydrate(ctx, &items[idx]); err != nil {
			return nil, 0, err
		}
	}
	return items, total, nil
}

func (r *ContentRepository) Get(ctx context.Context, id string) (*model.ContentItem, error) {
	var i model.ContentItem
	err := r.db.QueryRowContext(ctx, `SELECT id,platform,content_format,external_post_id,canonical_url,author_handle,original_text,cleaned_original_text,published_at,source_query,status,created_at,updated_at FROM content_items WHERE id=$1`, id).Scan(&i.ID, &i.Platform, &i.ContentFormat, &i.ExternalPostID, &i.CanonicalURL, &i.AuthorHandle, &i.OriginalText, &i.CleanedOriginalText, &i.PublishedAt, &i.SourceQuery, &i.Status, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if err := r.hydrate(ctx, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *ContentRepository) hydrate(ctx context.Context, item *model.ContentItem) error {
	var media []byte
	if err := r.db.QueryRowContext(ctx, `SELECT media FROM content_items WHERE id=$1`, item.ID).Scan(&media); err != nil {
		return err
	}
	if len(media) > 0 && string(media) != "null" {
		if err := json.Unmarshal(media, &item.Media); err != nil {
			return err
		}
	}
	rows, err := r.db.QueryContext(ctx, `SELECT cn.id,cn.name,cn.slug,cn.description,cn.status,cn.created_at,cn.updated_at FROM content_niches cn JOIN content_item_niches cin ON cin.content_niche_id=cn.id WHERE cin.content_item_id=$1 ORDER BY cn.name`, item.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var n model.ContentNiche
		if err := rows.Scan(&n.ID, &n.Name, &n.Slug, &n.Description, &n.Status, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return err
		}
		item.Niches = append(item.Niches, n)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT n.id,n.name,n.created_at FROM niches n JOIN content_item_product_types cipt ON cipt.product_type_id=n.id WHERE cipt.content_item_id=$1 ORDER BY n.name`, item.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var n model.Niche
		if err := rows.Scan(&n.ID, &n.Name, &n.CreatedAt); err != nil {
			return err
		}
		item.ProductTypes = append(item.ProductTypes, n)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	var stats model.ContentStats
	err = r.db.QueryRowContext(ctx, `SELECT like_count,repost_count,reply_count,bookmark_count,view_count,captured_at FROM content_stat_snapshots WHERE content_item_id=$1 ORDER BY captured_at DESC LIMIT 1`, item.ID).Scan(&stats.LikeCount, &stats.RepostCount, &stats.ReplyCount, &stats.BookmarkCount, &stats.ViewCount, &stats.CapturedAt)
	if err == nil {
		item.LatestStats = &stats
	} else if err != sql.ErrNoRows {
		return err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT id,content_item_id,name,text,source,model,position,created_at,updated_at FROM content_item_variants WHERE content_item_id=$1 ORDER BY position,id`, item.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var v model.ContentVariant
		if err := rows.Scan(&v.ID, &v.ContentItemID, &v.Name, &v.Text, &v.Source, &v.Model, &v.Position, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return err
		}
		item.Variants = append(item.Variants, v)
	}
	return rows.Err()
}

type ContentUpdate struct {
	Platform            string
	ContentFormat       string
	ExternalPostID      string
	CanonicalURL        string
	AuthorHandle        string
	OriginalText        string
	CleanedOriginalText string
	Media               []string
	PublishedAt         *time.Time
	SourceQuery         string
	Status              string
}

func (r *ContentRepository) Update(ctx context.Context, id string, item ContentUpdate, nicheIDs, productTypeIDs []string, stats *model.ContentStats) (*model.ContentItem, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	media, err := json.Marshal(item.Media)
	if err != nil {
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE content_items SET platform=$1,content_format=$2,external_post_id=$3,canonical_url=$4,author_handle=$5,original_text=$6,cleaned_original_text=COALESCE(NULLIF($7,''),cleaned_original_text),media=$8::jsonb,published_at=$9,source_query=$10,status=$11,updated_at=CURRENT_TIMESTAMP WHERE id=$12`, item.Platform, item.ContentFormat, item.ExternalPostID, item.CanonicalURL, item.AuthorHandle, item.OriginalText, item.CleanedOriginalText, string(media), item.PublishedAt, item.SourceQuery, item.Status, id)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, sql.ErrNoRows
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM content_item_niches WHERE content_item_id=$1`, id); err != nil {
		return nil, err
	}
	for _, nid := range nicheIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_item_niches(content_item_id,content_niche_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, nid); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM content_item_product_types WHERE content_item_id=$1`, id); err != nil {
		return nil, err
	}
	for _, tid := range productTypeIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_item_product_types(content_item_id,product_type_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, tid); err != nil {
			return nil, err
		}
	}
	if stats != nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO content_stat_snapshots(content_item_id,like_count,repost_count,reply_count,bookmark_count,view_count) VALUES($1,$2,$3,$4,$5,$6)`, id, stats.LikeCount, stats.RepostCount, stats.ReplyCount, stats.BookmarkCount, stats.ViewCount); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *ContentRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM content_items WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ContentRepository) CreateVariant(ctx context.Context, itemID string, v model.ContentVariant) (*model.ContentVariant, error) {
	var out model.ContentVariant
	err := r.db.QueryRowContext(ctx, `INSERT INTO content_item_variants(content_item_id,name,text,source,model,position) VALUES($1,$2,$3,$4,$5,COALESCE($6,(SELECT COALESCE(MAX(position),0)+1 FROM content_item_variants WHERE content_item_id=$1))) RETURNING id,content_item_id,name,text,source,model,position,created_at,updated_at`, itemID, v.Name, v.Text, v.Source, v.Model, v.Position).Scan(&out.ID, &out.ContentItemID, &out.Name, &out.Text, &out.Source, &out.Model, &out.Position, &out.CreatedAt, &out.UpdatedAt)
	return &out, err
}
func (r *ContentRepository) UpdateVariant(ctx context.Context, id string, v model.ContentVariant) (*model.ContentVariant, error) {
	var out model.ContentVariant
	err := r.db.QueryRowContext(ctx, `UPDATE content_item_variants SET name=$1,text=$2,source=$3,model=$4,position=$5,updated_at=CURRENT_TIMESTAMP WHERE id=$6 RETURNING id,content_item_id,name,text,source,model,position,created_at,updated_at`, v.Name, v.Text, v.Source, v.Model, v.Position, id).Scan(&out.ID, &out.ContentItemID, &out.Name, &out.Text, &out.Source, &out.Model, &out.Position, &out.CreatedAt, &out.UpdatedAt)
	return &out, err
}
func (r *ContentRepository) DeleteVariant(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM content_item_variants WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
