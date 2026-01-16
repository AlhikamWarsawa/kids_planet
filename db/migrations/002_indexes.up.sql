-- USERS

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
    ON users (email);

CREATE INDEX IF NOT EXISTS idx_users_role_status
    ON users (role, status);

-- PLAYERS

CREATE INDEX IF NOT EXISTS idx_players_user_id
    ON players (user_id);

-- EDUCATION_CATEGORIES

CREATE UNIQUE INDEX IF NOT EXISTS idx_education_categories_name
    ON education_categories (name);

-- AGE_CATEGORIES

CREATE UNIQUE INDEX IF NOT EXISTS idx_age_categories_label
    ON age_categories (label);

CREATE INDEX IF NOT EXISTS idx_age_categories_range
    ON age_categories (min_age, max_age);

-- GAMES

CREATE UNIQUE INDEX IF NOT EXISTS idx_games_slug
    ON games (slug);

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

CREATE UNIQUE INDEX IF NOT EXISTS idx_gec_game_edu_unique
    ON game_education_categories (game_id, education_category_id);

-- ANALYTICS_EVENTS
-- CREATE INDEX IF NOT EXISTS idx_analytics_events_game_id_created_at
--     ON analytics_events (game_id, created_at DESC);
--
-- CREATE INDEX IF NOT EXISTS idx_analytics_events_player_id_created_at
--     ON analytics_events (player_id, created_at DESC);
--
-- CREATE INDEX IF NOT EXISTS idx_analytics_events_event_type
--     ON analytics_events (event_type);

-- LEADERBOARD

-- CREATE INDEX IF NOT EXISTS idx_leaderboard_game_id_recorded_at
--     ON leaderboard (game_id, recorded_at DESC);
--
-- CREATE INDEX IF NOT EXISTS idx_leaderboard_player_id_recorded_at
--     ON leaderboard (player_id, recorded_at DESC);
--
-- CREATE INDEX IF NOT EXISTS idx_leaderboard_game_id_score
--     ON leaderboard (game_id, score DESC);
