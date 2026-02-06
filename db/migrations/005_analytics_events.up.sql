CREATE TABLE analytics_events
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    session_id  TEXT        NOT NULL,
    game_id     BIGINT      NOT NULL,
    event_name  TEXT        NOT NULL,
    event_data  JSONB,
    ip          TEXT,
    user_agent  TEXT,

    CONSTRAINT fk_analytics_events_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_session_id
    ON analytics_events (session_id);

CREATE INDEX IF NOT EXISTS idx_analytics_events_game_id
    ON analytics_events (game_id);

CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at
    ON analytics_events (created_at DESC);
