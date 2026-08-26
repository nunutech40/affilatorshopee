CREATE TABLE IF NOT EXISTS post_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    platform VARCHAR(30) NOT NULL DEFAULT 'x' CHECK (platform = 'x'),
    caption TEXT NOT NULL,
    hashtags TEXT[],
    notes TEXT,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (hashtags IS NULL OR cardinality(hashtags) <= 3)
);

CREATE INDEX IF NOT EXISTS idx_post_logs_product_id ON post_logs(product_id);
CREATE INDEX IF NOT EXISTS idx_post_logs_posted_at ON post_logs(posted_at DESC);
