ALTER TABLE products DROP CONSTRAINT IF EXISTS products_content_model_check;
ALTER TABLE products ADD CONSTRAINT products_content_model_check
    CHECK (content_model IN ('capture', 'cheap', 'trending', 'branded', 'curated'));

CREATE TABLE IF NOT EXISTS product_tracking_archive (
    product_id UUID PRIMARY KEY,
    shopee_link TEXT NOT NULL DEFAULT '',
    tracking_tag VARCHAR(64) NOT NULL UNIQUE,
    product_name VARCHAR(255),
    content_model VARCHAR(20),
    deleted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    click_count INTEGER NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMPTZ,
    sales_count INTEGER NOT NULL DEFAULT 0,
    pending_sales_count INTEGER NOT NULL DEFAULT 0,
    commission_total BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_product_tracking_archive_tag ON product_tracking_archive(tracking_tag);
CREATE INDEX IF NOT EXISTS idx_product_tracking_archive_clicks ON product_tracking_archive(click_count DESC);
