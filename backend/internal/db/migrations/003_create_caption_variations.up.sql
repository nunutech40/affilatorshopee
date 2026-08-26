CREATE TABLE IF NOT EXISTS caption_variations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    template VARCHAR(50) NOT NULL,
    caption TEXT NOT NULL,
    hashtags TEXT[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (cardinality(hashtags) <= 3)
);

CREATE INDEX IF NOT EXISTS idx_caption_variations_product_id ON caption_variations(product_id);

DROP TRIGGER IF EXISTS caption_variations_set_updated_at ON caption_variations;
CREATE TRIGGER caption_variations_set_updated_at
BEFORE UPDATE ON caption_variations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
