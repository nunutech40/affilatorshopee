INSERT INTO product_niches (product_id, niche_id)
SELECT pn.product_id, buku.id
FROM product_niches pn
JOIN niches pengembangan ON pengembangan.id = pn.niche_id AND pengembangan.name = 'Pengembangan Diri'
JOIN niches buku ON buku.name = 'Buku'
ON CONFLICT DO NOTHING;

DELETE FROM niches WHERE name = 'Pengembangan Diri';

UPDATE niches
SET name = 'Buku & Pengembangan Diri'
WHERE name = 'Buku';
