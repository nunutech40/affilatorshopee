CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    raw_text TEXT NOT NULL,
    product_name VARCHAR(255),
    shopee_link TEXT NOT NULL,
    image_url TEXT,
    image_urls TEXT[],
    video_url TEXT,
    normal_price INTEGER CHECK (normal_price IS NULL OR normal_price >= 0),
    sale_price INTEGER CHECK (sale_price IS NULL OR sale_price >= 0),
    discount_percent INTEGER CHECK (discount_percent IS NULL OR discount_percent BETWEEN 0 AND 100),
    rating NUMERIC(2,1) CHECK (rating IS NULL OR rating BETWEEN 0 AND 5),
    sold_count VARCHAR(50),
    review_count VARCHAR(50),
    keyword VARCHAR(255),
    problem VARCHAR(255),
    cluster VARCHAR(100),
    content_model VARCHAR(20) CHECK (content_model IN ('capture', 'cheap')),
    capture_angle VARCHAR(20) CHECK (capture_angle IN ('search', 'reply', 'trend', 'problem')),
    CHECK (capture_angle IS NULL OR content_model = 'capture'),
    benefit_1 VARCHAR(255),
    benefit_2 VARCHAR(255),
    benefit_3 VARCHAR(255),
    urgency VARCHAR(255),
    caption_template VARCHAR(50) NOT NULL DEFAULT 'direct_product'
        CHECK (caption_template IN ('direct_product', 'keyword_recommendation', 'problem_specific', 'cheap_value')),
    hashtag_pool TEXT[],
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'raw'
        CHECK (status IN ('raw', 'reformatted', 'ready')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (sale_price IS NULL OR normal_price IS NULL OR sale_price <= normal_price),
    CHECK (hashtag_pool IS NULL OR cardinality(hashtag_pool) <= 3)
);

CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_cluster ON products(cluster);
CREATE INDEX IF NOT EXISTS idx_products_content_model ON products(content_model);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS products_set_updated_at ON products;
CREATE TRIGGER products_set_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
