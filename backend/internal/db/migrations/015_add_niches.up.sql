CREATE TABLE IF NOT EXISTS niches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS product_niches (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    niche_id UUID NOT NULL REFERENCES niches(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, niche_id)
);

CREATE INDEX IF NOT EXISTS idx_product_niches_niche_id ON product_niches(niche_id);

INSERT INTO niches (name) VALUES
    ('Buku & Pengembangan Diri'),
    ('Fashion Pria'),
    ('Fashion Wanita & Hijab'),
    ('Sepatu'),
    ('Elektronik & Gadget'),
    ('Rumah Tangga'),
    ('Bayi & Anak'),
    ('Makanan & Minuman'),
    ('Peralatan Konten'),
    ('Bisnis & Produktivitas')
ON CONFLICT (name) DO NOTHING;
