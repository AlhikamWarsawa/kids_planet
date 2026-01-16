INSERT INTO age_categories (label, min_age, max_age)
VALUES
    ('3+', 3, 4),
    ('5+', 5, 6),
    ('7+', 7, 9),
    ('10+', 10, 12)
ON CONFLICT (label) DO NOTHING;
