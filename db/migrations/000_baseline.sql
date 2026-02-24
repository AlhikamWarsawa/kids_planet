-- ENUM TYPES
DO $$
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
            CREATE TYPE user_role AS ENUM ('admin', 'player');
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_status') THEN
            CREATE TYPE user_status AS ENUM ('active', 'inactive');
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'game_status') THEN
            CREATE TYPE game_status AS ENUM ('draft', 'active', 'archived');
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'game_difficulty') THEN
            CREATE TYPE game_difficulty AS ENUM ('easy', 'medium', 'hard');
        END IF;
    END$$;

CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- USERS
CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    public_id     UUID,
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    -- If you use citext: email CITEXT NOT NULL UNIQUE
    password_hash VARCHAR(255),
    pin_hash      VARCHAR(255),
    role          user_role    NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_users_player_auth_fields
        CHECK (role <> 'player' OR (public_id IS NOT NULL AND pin_hash IS NOT NULL))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_public_id_unique
    ON users (public_id);

CREATE INDEX IF NOT EXISTS idx_users_role_status
    ON users (role, status);

CREATE INDEX IF NOT EXISTS idx_users_role
    ON users (role);

DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;
CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- PLAYERS
CREATE TABLE IF NOT EXISTS players
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT UNIQUE, -- allow NULL if you really want orphan players
    nickname   VARCHAR(100) NOT NULL,
    age        INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_players_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE SET NULL,

    CONSTRAINT ck_players_age_non_negative
        CHECK (age IS NULL OR age >= 0)
);

-- EDUCATION CATEGORIES
CREATE TABLE IF NOT EXISTS education_categories
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    icon       VARCHAR(100),
    color      VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- AGE CATEGORIES
CREATE TABLE IF NOT EXISTS age_categories
(
    id         BIGSERIAL PRIMARY KEY,
    label      VARCHAR(50) NOT NULL UNIQUE,
    min_age    INT         NOT NULL,
    max_age    INT         NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT ck_age_categories_valid_range
        CHECK (min_age >= 0 AND max_age >= min_age)
);

CREATE INDEX IF NOT EXISTS idx_age_categories_range
    ON age_categories (min_age, max_age);

-- GAMES
CREATE TABLE IF NOT EXISTS games
(
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(150) NOT NULL,
    slug            VARCHAR(150) NOT NULL UNIQUE,
    description     TEXT,
    thumbnail       VARCHAR(255),
    game_url        VARCHAR(255),
    difficulty      game_difficulty,
    age_category_id BIGINT       NOT NULL,
    free            BOOLEAN      NOT NULL DEFAULT TRUE,
    status          game_status  NOT NULL DEFAULT 'draft',
    created_by      BIGINT       NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_games_age_category
        FOREIGN KEY (age_category_id)
            REFERENCES age_categories (id)
            ON DELETE RESTRICT,

    CONSTRAINT fk_games_created_by
        FOREIGN KEY (created_by)
            REFERENCES users (id)
            ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_games_status
    ON games (status);

CREATE INDEX IF NOT EXISTS idx_games_age_category_id
    ON games (age_category_id);

CREATE INDEX IF NOT EXISTS idx_games_created_by
    ON games (created_by);

CREATE INDEX IF NOT EXISTS idx_games_created_at
    ON games (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_games_active
    ON games (status)
    WHERE status = 'active';

DROP TRIGGER IF EXISTS trg_games_set_updated_at ON games;
CREATE TRIGGER trg_games_set_updated_at
    BEFORE UPDATE ON games
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- GAME EDUCATION CATEGORIES
CREATE TABLE IF NOT EXISTS game_education_categories
(
    id                    BIGSERIAL PRIMARY KEY,
    game_id               BIGINT NOT NULL,
    education_category_id BIGINT NOT NULL,

    CONSTRAINT uq_game_education UNIQUE (game_id, education_category_id),

    CONSTRAINT fk_gec_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_gec_education
        FOREIGN KEY (education_category_id)
            REFERENCES education_categories (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_gec_game_id
    ON game_education_categories (game_id);

CREATE INDEX IF NOT EXISTS idx_gec_education_category_id
    ON game_education_categories (education_category_id);

-- SESSIONS (make them real)
CREATE TABLE IF NOT EXISTS sessions
(
    id                BIGSERIAL PRIMARY KEY,
    game_id            BIGINT       NOT NULL,
    client_session_id  TEXT,        -- optional: if your frontend generates a string/uuid session id
    started_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_sessions_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_started_at
    ON sessions (started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_game_id_started_at
    ON sessions (game_id, started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_client_session_id
    ON sessions (client_session_id);

-- LEADERBOARD SUBMISSIONS
CREATE TABLE IF NOT EXISTS leaderboard_submissions
(
    id                  BIGSERIAL PRIMARY KEY,
    game_id              BIGINT       NOT NULL,
    player_id            BIGINT,
    session_id           BIGINT,      -- FIX: FK to sessions.id
    member               TEXT,
    score                INT          NOT NULL,
    ip_hash              TEXT,
    user_agent_hash      TEXT,
    flagged              BOOLEAN      NOT NULL DEFAULT FALSE,
    flag_reason          TEXT,
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    removed_by_admin_id  BIGINT,
    removed_at           TIMESTAMPTZ,

    CONSTRAINT fk_leaderboard_submissions_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_leaderboard_submissions_player
        FOREIGN KEY (player_id)
            REFERENCES players (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_leaderboard_submissions_session
        FOREIGN KEY (session_id)
            REFERENCES sessions (id)
            ON DELETE SET NULL,

    CONSTRAINT fk_leaderboard_submissions_removed_by_admin
        FOREIGN KEY (removed_by_admin_id)
            REFERENCES users (id)
            ON DELETE SET NULL,

    CONSTRAINT ck_leaderboard_submissions_score_non_negative
        CHECK (score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_lb_game_created_at
    ON leaderboard_submissions (game_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_lb_flagged_created_at
    ON leaderboard_submissions (flagged, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_lb_removed_at
    ON leaderboard_submissions (removed_at DESC)
    WHERE removed_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_lb_game_score
    ON leaderboard_submissions (game_id, score DESC);

CREATE INDEX IF NOT EXISTS idx_lb_flagged
    ON leaderboard_submissions (flagged)
    WHERE flagged = TRUE;

DROP TRIGGER IF EXISTS trg_lb_set_updated_at ON leaderboard_submissions;
CREATE TRIGGER trg_lb_set_updated_at
    BEFORE UPDATE ON leaderboard_submissions
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- ANALYTICS EVENTS
CREATE TABLE IF NOT EXISTS analytics_events
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    session_id BIGINT,     -- FIX: FK to sessions.id
    game_id    BIGINT      NOT NULL,
    event_name TEXT        NOT NULL,
    event_data JSONB,
    ip         TEXT,
    user_agent TEXT,

    CONSTRAINT fk_analytics_events_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_analytics_events_session
        FOREIGN KEY (session_id)
            REFERENCES sessions (id)
            ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_session_id
    ON analytics_events (session_id);

CREATE INDEX IF NOT EXISTS idx_analytics_events_game_id
    ON analytics_events (game_id);

CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at
    ON analytics_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_analytics_events_event_name_created_at_game_id
    ON analytics_events (event_name, created_at DESC, game_id);