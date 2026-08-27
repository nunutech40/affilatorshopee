ALTER TABLE products DROP CONSTRAINT IF EXISTS products_source_category_check;
ALTER TABLE products ADD CONSTRAINT products_source_category_check CHECK (source_category IN ('raw_text', 'import_x', 'scrape_shopee'));
