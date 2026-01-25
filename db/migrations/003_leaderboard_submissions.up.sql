CREATE TABLE leaderboard_submissions
(
    id              BIGSERIAL PRIMARY KEY,
    game_id          BIGINT  NOT NULL,
    player_id        BIGINT,
    session_id       TEXT,
    score            INT     NOT NULL,
    ip_hash          TEXT,
    user_agent_hash  TEXT,
    flagged          BOOLEAN NOT NULL DEFAULT FALSE,
    flag_reason      TEXT,
    created_at       TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_leaderboard_submissions_game
        FOREIGN KEY (game_id)
            REFERENCES games (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_leaderboard_submissions_player
        FOREIGN KEY (player_id)
            REFERENCES players (id)
            ON DELETE SET NULL,

    CONSTRAINT ck_leaderboard_submissions_score_non_negative
        CHECK (score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_leaderboard_submissions_game_created_at
    ON leaderboard_submissions (game_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_leaderboard_submissions_flagged_created_at
    ON leaderboard_submissions (flagged, created_at DESC);