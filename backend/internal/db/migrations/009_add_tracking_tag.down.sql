DROP INDEX IF EXISTS idx_products_tracking_tag;
ALTER TABLE products DROP COLUMN IF EXISTS tracking_tag;
