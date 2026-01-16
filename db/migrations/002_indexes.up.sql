-- USERS
CREATE INDEX IF NOT EXISTS idx_users_role_status
    ON users (role, status);

-- AGE_CATEGORIES
CREATE INDEX IF NOT EXISTS idx_age_categories_range
    ON age_categories (min_age, max_age);

-- GAMES
CREATE INDEX IF NOT EXISTS idx_games_status
    ON games (status);

CREATE INDEX IF NOT EXISTS idx_games_age_category_id
    ON games (age_category_id);

CREATE INDEX IF NOT EXISTS idx_games_created_by
    ON games (created_by);

CREATE INDEX IF NOT EXISTS idx_games_created_at
    ON games (created_at DESC);

-- GAME_EDUCATION_CATEGORIES
CREATE INDEX IF NOT EXISTS idx_gec_game_id
    ON game_education_categories (game_id);

CREATE INDEX IF NOT EXISTS idx_gec_education_category_id
    ON game_education_categories (education_category_id);