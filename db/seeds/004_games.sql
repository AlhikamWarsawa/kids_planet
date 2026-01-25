INSERT INTO games (
    title,
    slug,
    description,
    difficulty,
    age_category_id,
    free,
    status,
    created_by
)
SELECT
    'Free Fire',
    'free-fire',
    'Shot Player booyah.',
    'hard',
    ac.id,
    TRUE,
    'active',
    u.id
FROM users u
         JOIN age_categories ac ON ac.label = '5+'
WHERE u.email = 'admin@kidsplanet.com'
ON CONFLICT (slug) DO NOTHING;
