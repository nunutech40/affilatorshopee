-- Shopee tracking tags accept alphanumeric characters only.
UPDATE products
SET tracking_tag = LEFT(regexp_replace(lower(tracking_tag), '[^a-z0-9]', '', 'g'), 64)
WHERE tracking_tag ~ '[^a-zA-Z0-9]';
