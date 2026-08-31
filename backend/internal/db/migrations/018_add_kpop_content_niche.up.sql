INSERT INTO content_niches (name, slug, description)
VALUES (
    'K-Pop & Fandom Korea',
    'k-pop-fandom-korea',
    'Berita, comeback, konser, fandom, fan project, chart, performance, merchandise, dan isu K-pop.'
)
ON CONFLICT (slug) DO NOTHING;
