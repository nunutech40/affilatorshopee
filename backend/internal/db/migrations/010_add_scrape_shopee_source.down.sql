ALTER TABLE products DROP CONSTRAINT IF EXISTS products_source_category_check;
UPDATE products SET source_category = 'raw_text' WHERE source_category = 'scrape_shopee';
ALTER TABLE products ADD CONSTRAINT products_source_category_check CHECK (source_category IN ('raw_text', 'import_x'));
