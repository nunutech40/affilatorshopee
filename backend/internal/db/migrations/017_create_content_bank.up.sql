CREATE TABLE IF NOT EXISTS content_niches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(120) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS content_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(20) NOT NULL DEFAULT 'x',
    external_post_id VARCHAR(150) NOT NULL DEFAULT '',
    canonical_url TEXT NOT NULL,
    author_handle VARCHAR(150) NOT NULL DEFAULT '',
    original_text TEXT NOT NULL,
    media JSONB NOT NULL DEFAULT '[]'::jsonb,
    published_at TIMESTAMPTZ,
    source_query TEXT NOT NULL DEFAULT '',
    status VARCHAR(30) NOT NULL DEFAULT 'discovered',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (platform, canonical_url)
);

CREATE TABLE IF NOT EXISTS content_item_niches (
    content_item_id UUID NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
    content_niche_id UUID NOT NULL REFERENCES content_niches(id) ON DELETE CASCADE,
    PRIMARY KEY (content_item_id, content_niche_id)
);

CREATE TABLE IF NOT EXISTS content_item_product_types (
    content_item_id UUID NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
    product_type_id UUID NOT NULL REFERENCES niches(id) ON DELETE CASCADE,
    PRIMARY KEY (content_item_id, product_type_id)
);

CREATE TABLE IF NOT EXISTS content_stat_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content_item_id UUID NOT NULL REFERENCES content_items(id) ON DELETE CASCADE,
    captured_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    like_count BIGINT NOT NULL DEFAULT 0,
    repost_count BIGINT NOT NULL DEFAULT 0,
    reply_count BIGINT NOT NULL DEFAULT 0,
    bookmark_count BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    raw_metrics JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_content_items_platform_status ON content_items(platform, status);
CREATE INDEX IF NOT EXISTS idx_content_item_niches_niche ON content_item_niches(content_niche_id);
CREATE INDEX IF NOT EXISTS idx_content_stats_item_captured ON content_stat_snapshots(content_item_id, captured_at DESC);

INSERT INTO content_niches (name, slug) VALUES
    ('Sukses & Kesuksesan', 'sukses-kesuksesan'),
    ('Fashion Pria', 'fashion-pria'),
    ('Hubungan / Relasi Pria Wanita', 'hubungan-relasi-pria-wanita'),
    ('Gym, Lari & Exercise', 'gym-lari-exercise'),
    ('Affiliate', 'affiliate')
ON CONFLICT (slug) DO NOTHING;
