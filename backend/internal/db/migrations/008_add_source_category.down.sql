ALTER TABLE products DROP CONSTRAINT IF EXISTS products_source_category_check;
ALTER TABLE products DROP COLUMN IF EXISTS source_category;
