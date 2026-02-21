-- MVP baseline seed (Day 28 freeze)
-- Minimal deterministic data for admin login + catalog/session/leaderboard/moderation flows.

-- ADMIN USER
INSERT INTO users (name, email, password_hash, role, status)
VALUES (
    'Admin',
    'admin@kidsplanet.com',
    '$2a$10$FuBgv/iPOnaSiQzbkEWWbOtlOz3bhsuYc5xVlkGFxVGEToro/qLQ2',
    'admin',
    'active'
)
ON CONFLICT (email) DO NOTHING;

-- AGE CATEGORIES
INSERT INTO age_categories (label, min_age, max_age)
VALUES
    ('3+', 3, 4),
    ('5+', 5, 6),
    ('7+', 7, 9),
    ('10+', 10, 12)
ON CONFLICT (label) DO NOTHING;

-- EDUCATION CATEGORIES
INSERT INTO education_categories (name, icon, color)
VALUES
    ('Math', 'calculator', '#4CAF50'),
    ('Reading', 'book', '#2196F3'),
    ('Logic', 'brain', '#9C27B0'),
    ('Memory', 'memory', '#FF9800'),
    ('Creativity', 'palette', '#E91E63')
ON CONFLICT (name) DO NOTHING;

-- GAMES (active and playable)
INSERT INTO games (
    title,
    slug,
    description,
    thumbnail,
    game_url,
    difficulty,
    age_category_id,
    free,
    status,
    created_by
)
SELECT
    'Math Sprint',
    'math-sprint',
    'Quick number challenges for kids.',
    '/img/math-sprint.png',
    '/games/math-sprint/current/index.html',
    'easy',
    ac.id,
    TRUE,
    'active',
    u.id
FROM users u
JOIN age_categories ac ON ac.label = '5+'
WHERE u.email = 'admin@kidsplanet.com'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO games (
    title,
    slug,
    description,
    thumbnail,
    game_url,
    difficulty,
    age_category_id,
    free,
    status,
    created_by
)
SELECT
    'Word Garden',
    'word-garden',
    'Vocabulary and reading mini-games.',
    '/img/word-garden.png',
    '/games/word-garden/current/index.html',
    'medium',
    ac.id,
    TRUE,
    'active',
    u.id
FROM users u
JOIN age_categories ac ON ac.label = '7+'
WHERE u.email = 'admin@kidsplanet.com'
ON CONFLICT (slug) DO NOTHING;

INSERT INTO games (
    title,
    slug,
    description,
    thumbnail,
    game_url,
    difficulty,
    age_category_id,
    free,
    status,
    created_by
)
SELECT
    'Logic Quest',
    'logic-quest',
    'Pattern and memory puzzles.',
    '/img/logic-quest.png',
    '/games/logic-quest/current/index.html',
    'hard',
    ac.id,
    TRUE,
    'active',
    u.id
FROM users u
JOIN age_categories ac ON ac.label = '10+'
WHERE u.email = 'admin@kidsplanet.com'
ON CONFLICT (slug) DO NOTHING;

-- GAME <-> EDUCATION CATEGORY LINKS
INSERT INTO game_education_categories (game_id, education_category_id)
SELECT g.id, ec.id
FROM games g
JOIN education_categories ec ON ec.name = 'Math'
WHERE g.slug = 'math-sprint'
ON CONFLICT (game_id, education_category_id) DO NOTHING;

INSERT INTO game_education_categories (game_id, education_category_id)
SELECT g.id, ec.id
FROM games g
JOIN education_categories ec ON ec.name = 'Reading'
WHERE g.slug = 'word-garden'
ON CONFLICT (game_id, education_category_id) DO NOTHING;

INSERT INTO game_education_categories (game_id, education_category_id)
SELECT g.id, ec.id
FROM games g
JOIN education_categories ec ON ec.name = 'Logic'
WHERE g.slug = 'logic-quest'
ON CONFLICT (game_id, education_category_id) DO NOTHING;

INSERT INTO game_education_categories (game_id, education_category_id)
SELECT g.id, ec.id
FROM games g
JOIN education_categories ec ON ec.name = 'Memory'
WHERE g.slug = 'logic-quest'
ON CONFLICT (game_id, education_category_id) DO NOTHING;
