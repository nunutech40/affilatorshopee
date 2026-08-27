ALTER TABLE products ADD COLUMN IF NOT EXISTS sales_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS pending_sales_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS commission_total BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS commission_events (
    event_id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    item_id TEXT,
    model_id TEXT,
    order_status TEXT NOT NULL,
    ordered_at TIMESTAMPTZ,
    tracking_tag TEXT,
    normalized_tag TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    commission_total BIGINT NOT NULL DEFAULT 0,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_commission_events_normalized_tag ON commission_events(normalized_tag);
