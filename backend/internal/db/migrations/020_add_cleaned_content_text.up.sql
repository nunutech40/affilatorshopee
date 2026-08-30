ALTER TABLE content_items ADD COLUMN IF NOT EXISTS cleaned_original_text TEXT NOT NULL DEFAULT '';
