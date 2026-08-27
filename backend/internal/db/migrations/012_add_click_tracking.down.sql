DROP TABLE IF EXISTS click_events;
ALTER TABLE products DROP COLUMN IF EXISTS last_clicked_at;
ALTER TABLE products DROP COLUMN IF EXISTS click_count;
