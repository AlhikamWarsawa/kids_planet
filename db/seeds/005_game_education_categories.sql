INSERT INTO game_education_categories (game_id, education_category_id)
SELECT
    g.id,
    ec.id
FROM games g
         JOIN education_categories ec
              ON ec.name IN ('Logic', 'Memory')
WHERE g.slug = 'color-match'
ON CONFLICT DO NOTHING;
