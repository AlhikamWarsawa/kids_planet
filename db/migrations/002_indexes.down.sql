-- LEADERBOARD
-- DROP INDEX IF EXISTS idx_leaderboard_game_id_score;
-- DROP INDEX IF EXISTS idx_leaderboard_player_id_recorded_at;
-- DROP INDEX IF EXISTS idx_leaderboard_game_id_recorded_at;

-- ANALYTICS_EVENTS
-- DROP INDEX IF EXISTS idx_analytics_events_event_type;
-- DROP INDEX IF EXISTS idx_analytics_events_player_id_created_at;
-- DROP INDEX IF EXISTS idx_analytics_events_game_id_created_at;

-- GAME_EDUCATION_CATEGORIES
DROP INDEX IF EXISTS idx_gec_game_edu_unique;
DROP INDEX IF EXISTS idx_gec_education_category_id;
DROP INDEX IF EXISTS idx_gec_game_id;

-- GAMES
DROP INDEX IF EXISTS idx_games_created_at;
DROP INDEX IF EXISTS idx_games_created_by;
DROP INDEX IF EXISTS idx_games_age_category_id;
DROP INDEX IF EXISTS idx_games_status;
DROP INDEX IF EXISTS idx_games_slug;

-- AGE_CATEGORIES
DROP INDEX IF EXISTS idx_age_categories_range;
DROP INDEX IF EXISTS idx_age_categories_label;

-- EDUCATION_CATEGORIES
DROP INDEX IF EXISTS idx_education_categories_name;

-- PLAYERS
DROP INDEX IF EXISTS idx_players_user_id;

-- USERS
DROP INDEX IF EXISTS idx_users_role_status;
DROP INDEX IF EXISTS idx_users_email;
