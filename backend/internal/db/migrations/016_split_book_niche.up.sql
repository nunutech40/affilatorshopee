UPDATE niches
SET name = 'Buku'
WHERE name = 'Buku & Pengembangan Diri';

INSERT INTO niches (name)
VALUES ('Pengembangan Diri')
ON CONFLICT (name) DO NOTHING;
