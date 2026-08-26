ALTER TABLE products ADD COLUMN IF NOT EXISTS tracking_tag VARCHAR(64);

UPDATE products
SET tracking_tag = 'p-' || substring(replace(id::text, '-', ''), 1, 12)
WHERE tracking_tag IS NULL OR tracking_tag = '';

ALTER TABLE products ALTER COLUMN tracking_tag SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_tracking_tag ON products(tracking_tag);
