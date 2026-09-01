DROP TABLE IF EXISTS product_tracking_archive;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_content_model_check;
ALTER TABLE products ADD CONSTRAINT products_content_model_check
    CHECK (content_model IN ('capture', 'cheap', 'trending', 'branded'));
