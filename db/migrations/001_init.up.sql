-- ENUM TYPES

CREATE TYPE user_role AS ENUM ('admin', 'player');
CREATE TYPE user_status AS ENUM ('active', 'inactive');
CREATE TYPE game_status AS ENUM ('draft', 'active', 'archived');
CREATE TYPE game_difficulty AS ENUM ('easy', 'medium', 'hard');

-- USERS

CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(150) NOT NULL,
    email         VARCHAR(150) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role          user_role    NOT NULL,
    status        user_status  NOT NULL DEFAULT 'active',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

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

-- EDUCATION_CATEGORIES

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
    created_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

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

-- GAME_EDUCATION_CATEGORIES

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