DROP TABLE IF EXISTS commission_events;
ALTER TABLE products DROP COLUMN IF EXISTS commission_total;
ALTER TABLE products DROP COLUMN IF EXISTS pending_sales_count;
ALTER TABLE products DROP COLUMN IF EXISTS sales_count;
