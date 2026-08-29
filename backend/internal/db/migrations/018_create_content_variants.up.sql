CREATE TABLE IF NOT EXISTS content_item_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_item_id UUID NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL DEFAULT '',
    text TEXT NOT NULL DEFAULT '',
    source VARCHAR(20) NOT NULL DEFAULT 'manual',
    model VARCHAR(200) NOT NULL DEFAULT '',
    position INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_content_variants_item ON content_item_variants(content_item_id, position, created_at);
