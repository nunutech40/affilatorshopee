package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/nunutech40/affilatorshopee/internal/model"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

type ProductListFilter struct {
	Cluster      string
	ContentModel string
	Status       string
	Search       string
	Page         int
	Limit        int
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	query := `INSERT INTO products (
		raw_text, reformatted_text, product_name, shopee_link, image_url, image_urls, video_url,
		normal_price, sale_price, discount_percent, rating, sold_count, review_count,
		keyword, problem, cluster, content_model, capture_angle,
		benefit_1, benefit_2, benefit_3, urgency, caption_template, hashtag_pool,
		notes, source_category, status
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
	RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		product.RawText, product.ReformattedText, product.ProductName, product.ShopeeLink, product.ImageURL,
		pq.Array(product.ImageURLs), product.VideoURL, product.NormalPrice,
		product.SalePrice, product.DiscountPercent, product.Rating, product.SoldCount,
		product.ReviewCount, product.Keyword, product.Problem, product.Cluster,
		product.ContentModel, product.CaptureAngle, product.Benefit1, product.Benefit2,
		product.Benefit3, product.Urgency, product.CaptionTemplate,
		pq.Array(product.HashtagPool), product.Notes, product.SourceCategory, product.Status,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*model.Product, error) {
	query := `SELECT
		p.id, p.raw_text, p.reformatted_text, p.product_name, p.shopee_link, p.image_url, p.image_urls,
		p.video_url, p.normal_price, p.sale_price, p.discount_percent, p.rating,
		p.sold_count, p.review_count, p.keyword, p.problem, p.cluster,
		p.content_model, p.capture_angle, p.benefit_1, p.benefit_2, p.benefit_3,
		p.urgency, p.caption_template, p.hashtag_pool, p.notes, p.source_category, p.status,
		p.created_at, p.updated_at,
		COALESCE((SELECT COUNT(*) FROM post_logs pl WHERE pl.product_id = p.id), 0) AS post_count,
		(SELECT MAX(pl.posted_at) FROM post_logs pl WHERE pl.product_id = p.id) AS last_posted_at
	FROM products p WHERE p.id = $1`

	return r.scanProduct(r.db.QueryRowContext(ctx, query, id))
}

func (r *ProductRepository) List(ctx context.Context, filter ProductListFilter) ([]model.Product, int, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if filter.Cluster != "" {
		where = append(where, fmt.Sprintf("p.cluster = $%d", argPos))
		args = append(args, filter.Cluster)
		argPos++
	}
	if filter.ContentModel != "" {
		where = append(where, fmt.Sprintf("p.content_model = $%d", argPos))
		args = append(args, filter.ContentModel)
		argPos++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("p.status = $%d", argPos))
		args = append(args, filter.Status)
		argPos++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(p.product_name ILIKE $%d OR p.keyword ILIKE $%d OR p.cluster ILIKE $%d OR p.raw_text ILIKE $%d)", argPos, argPos, argPos, argPos))
		args = append(args, "%"+filter.Search+"%")
		argPos++
	}

	whereClause := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT
		p.id, p.raw_text, p.reformatted_text, p.product_name, p.shopee_link, p.image_url, p.image_urls,
		p.video_url, p.normal_price, p.sale_price, p.discount_percent, p.rating,
		p.sold_count, p.review_count, p.keyword, p.problem, p.cluster,
		p.content_model, p.capture_angle, p.benefit_1, p.benefit_2, p.benefit_3,
		p.urgency, p.caption_template, p.hashtag_pool, p.notes, p.source_category, p.status,
		p.created_at, p.updated_at,
		COALESCE((SELECT COUNT(*) FROM post_logs pl WHERE pl.product_id = p.id), 0) AS post_count,
		(SELECT MAX(pl.posted_at) FROM post_logs pl WHERE pl.product_id = p.id) AS last_posted_at
	FROM products p WHERE %s ORDER BY p.created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argPos, argPos+1)
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := []model.Product{}
	for rows.Next() {
		product, err := r.scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, *product)
	}
	return products, total, rows.Err()
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	return r.update(ctx, product, "")
}

func (r *ProductRepository) UpdateReformatted(ctx context.Context, product *model.Product) error {
	return r.update(ctx, product, "raw")
}

func (r *ProductRepository) update(ctx context.Context, product *model.Product, requiredStatus string) error {
	query := `UPDATE products SET
		reformatted_text=$2, product_name=$3, shopee_link=$4, image_url=$5, image_urls=$6, video_url=$7,
		normal_price=$8, sale_price=$9, discount_percent=$10, rating=$11,
		sold_count=$12, review_count=$13, keyword=$14, problem=$15, cluster=$16,
		content_model=$17, capture_angle=$18, benefit_1=$19, benefit_2=$20, benefit_3=$21,
		urgency=$22, caption_template=$23, hashtag_pool=$24, notes=$25, source_category=$26, status=$27
	WHERE id=$1`
	args := []interface{}{
		product.ID, product.ReformattedText, product.ProductName, product.ShopeeLink, product.ImageURL,
		pq.Array(product.ImageURLs), product.VideoURL, product.NormalPrice,
		product.SalePrice, product.DiscountPercent, product.Rating, product.SoldCount,
		product.ReviewCount, product.Keyword, product.Problem, product.Cluster,
		product.ContentModel, product.CaptureAngle, product.Benefit1, product.Benefit2,
		product.Benefit3, product.Urgency, product.CaptionTemplate,
		pq.Array(product.HashtagPool), product.Notes, product.SourceCategory, product.Status,
	}
	if requiredStatus != "" {
		query = query + " AND status=$27"
		args = append(args, requiredStatus)
	}
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	if requiredStatus != "" {
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

func (r *ProductRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE products SET status=$2 WHERE id=$1`, id, status)
	return err
}

func (r *ProductRepository) scanProduct(row interface {
	Scan(dest ...interface{}) error
}) (*model.Product, error) {
	var p model.Product
	err := row.Scan(
		&p.ID, &p.RawText, &p.ReformattedText, &p.ProductName, &p.ShopeeLink, &p.ImageURL,
		pq.Array(&p.ImageURLs), &p.VideoURL, &p.NormalPrice, &p.SalePrice,
		&p.DiscountPercent, &p.Rating, &p.SoldCount, &p.ReviewCount,
		&p.Keyword, &p.Problem, &p.Cluster, &p.ContentModel, &p.CaptureAngle,
		&p.Benefit1, &p.Benefit2, &p.Benefit3, &p.Urgency, &p.CaptionTemplate,
		pq.Array(&p.HashtagPool), &p.Notes, &p.SourceCategory, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		&p.PostCount, &p.LastPostedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
