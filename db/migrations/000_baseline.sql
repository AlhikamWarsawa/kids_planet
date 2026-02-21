-- ENUM TYPES

CREATE TYPE user_role AS ENUM ('admin', 'player');
CREATE TYPE user_status AS ENUM ('active', 'inactive');
CREATE TYPE game_status AS ENUM ('draft', 'active', 'archived');
CREATE TYPE game_difficulty AS ENUM ('easy', 'medium', 'hard');

-- USERS

CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    public_id     UUID UNIQUE,
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    pin_hash      VARCHAR(255),
    role          user_role    NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_users_player_auth_fields
        CHECK (role <> 'player' OR (public_id IS NOT NULL AND pin_hash IS NOT NULL))
);

CREATE UNIQUE INDEX idx_users_public_id_unique
    ON users (public_id)
    WHERE public_id IS NOT NULL;

CREATE INDEX idx_users_role_status
    ON users (role, status);

-- PLAYERS

CREATE TABLE players
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT UNIQUE,
    nickname   VARCHAR(100) NOT NULL,
    age        INT,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_players_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE SET NULL
);

-- EDUCATION CATEGORIES

CREATE TABLE education_categories
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL UNIQUE,
    icon       VARCHAR(100),
    color      VARCHAR(20),
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- AGE CATEGORIES

CREATE TABLE age_categories
(
    id         BIGSERIAL PRIMARY KEY,
    label      VARCHAR(50) NOT NULL UNIQUE,
    min_age    INT         NOT NULL,
    max_age    INT         NOT NULL,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_age_categories_valid_range
        CHECK (min_age >= 0 AND max_age >= min_age)
);

CREATE INDEX idx_age_categories_range
    ON age_categories (min_age, max_age);

-- GAMES

CREATE TABLE games
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
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_games_age_category
        FOREIGN KEY (age_category_id)
            REFERENCES age_categories (id)
            ON DELETE RESTRICT,
    CONSTRAINT fk_games_created_by
        FOREIGN KEY (created_by)
            REFERENCES users (id)
            ON DELETE RESTRICT
);

CREATE INDEX idx_games_status
    ON games (status);

CREATE INDEX idx_games_age_category_id
    ON games (age_category_id);

CREATE INDEX idx_games_created_by
    ON games (created_by);

CREATE INDEX idx_games_created_at
    ON games (created_at DESC);

-- GAME EDUCATION CATEGORIES

CREATE TABLE game_education_categories
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

CREATE INDEX idx_gec_game_id
    ON game_education_categories (game_id);

CREATE INDEX idx_gec_education_category_id
    ON game_education_categories (education_category_id);

-- LEADERBOARD SUBMISSIONS

CREATE TABLE leaderboard_submissions
(
    id                  BIGSERIAL PRIMARY KEY,
    game_id             BIGINT    NOT NULL,
    player_id           BIGINT,
    session_id          TEXT,
    member              TEXT,
    score               INT       NOT NULL,
    ip_hash             TEXT,
    user_agent_hash     TEXT,
    flagged             BOOLEAN   NOT NULL DEFAULT FALSE,
    flag_reason         TEXT,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    removed_by_admin_id BIGINT,
    removed_at          TIMESTAMP,
    CONSTRAINT fk_leaderboard_submissions_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_leaderboard_submissions_player
        FOREIGN KEY (player_id)
            REFERENCES players (id)
            ON DELETE SET NULL,
    CONSTRAINT fk_leaderboard_submissions_removed_by_admin
        FOREIGN KEY (removed_by_admin_id)
            REFERENCES users (id)
            ON DELETE SET NULL,
    CONSTRAINT ck_leaderboard_submissions_score_non_negative
        CHECK (score >= 0)
);

CREATE INDEX idx_leaderboard_submissions_game_created_at
    ON leaderboard_submissions (game_id, created_at DESC);

CREATE INDEX idx_leaderboard_submissions_flagged_created_at
    ON leaderboard_submissions (flagged, created_at DESC);

CREATE INDEX idx_leaderboard_submissions_removed_at
    ON leaderboard_submissions (removed_at DESC);

-- ANALYTICS EVENTS

CREATE TABLE analytics_events
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    session_id TEXT        NOT NULL,
    game_id    BIGINT      NOT NULL,
    event_name TEXT        NOT NULL,
    event_data JSONB,
    ip         TEXT,
    user_agent TEXT,
    CONSTRAINT fk_analytics_events_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_analytics_events_session_id
    ON analytics_events (session_id);

CREATE INDEX idx_analytics_events_game_id
    ON analytics_events (game_id);

CREATE INDEX idx_analytics_events_created_at
    ON analytics_events (created_at DESC);

-- SESSIONS

CREATE TABLE sessions
(
    id         BIGSERIAL PRIMARY KEY,
    game_id    BIGINT    NOT NULL,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_sessions_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_sessions_started_at
    ON sessions (started_at DESC);

CREATE INDEX idx_sessions_game_id_started_at
    ON sessions (game_id, started_at DESC);