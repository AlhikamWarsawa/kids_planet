-- GAME_EDUCATION_CATEGORIES
DROP INDEX IF EXISTS idx_gec_education_category_id;
DROP INDEX IF EXISTS idx_gec_game_id;

-- GAMES
DROP INDEX IF EXISTS idx_games_created_at;
DROP INDEX IF EXISTS idx_games_created_by;
DROP INDEX IF EXISTS idx_games_age_category_id;
DROP INDEX IF EXISTS idx_games_status;

-- AGE_CATEGORIES
DROP INDEX IF EXISTS idx_age_categories_range;

-- USERS
DROP INDEX IF EXISTS idx_users_role_status;
