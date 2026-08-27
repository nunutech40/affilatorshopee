ALTER TABLE products ADD COLUMN IF NOT EXISTS click_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS last_clicked_at TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS click_events (
    click_id TEXT PRIMARY KEY,
    clicked_at TIMESTAMPTZ NOT NULL,
    region TEXT,
    tracking_tag TEXT NOT NULL,
    normalized_tag TEXT NOT NULL,
    referrer TEXT,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_click_events_normalized_tag ON click_events(normalized_tag);
CREATE INDEX IF NOT EXISTS idx_click_events_clicked_at ON click_events(clicked_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_click_count ON products(click_count DESC);
