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
    'Color Match',
    'color-match',
    'Match colors to improve memory and logic skills.',
    'easy',
    ac.id,
    TRUE,
    'draft',
    u.id
FROM users u
         JOIN age_categories ac ON ac.label = '5+'
WHERE u.email = 'admin@kidsplanet.com'
ON CONFLICT (slug) DO NOTHING;
