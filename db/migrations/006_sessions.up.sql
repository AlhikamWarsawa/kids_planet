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

CREATE INDEX IF NOT EXISTS idx_sessions_started_at
    ON sessions (started_at DESC);

CREATE INDEX IF NOT EXISTS idx_sessions_game_id_started_at
    ON sessions (game_id, started_at DESC);
