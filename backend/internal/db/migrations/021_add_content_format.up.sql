ALTER TABLE content_items ADD COLUMN IF NOT EXISTS content_format VARCHAR(20) NOT NULL DEFAULT 'post';
UPDATE content_items SET content_format = CASE WHEN original_text ~* '(^|\n)\s*Post\s+[0-9]+' OR length(original_text) > 280 THEN 'thread' ELSE 'post' END WHERE content_format = 'post';
