ALTER TABLE leaderboard_submissions
    ADD COLUMN IF NOT EXISTS removed_by_admin_id BIGINT,
    ADD COLUMN IF NOT EXISTS removed_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS member TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_leaderboard_submissions_removed_by_admin'
    ) THEN
        ALTER TABLE leaderboard_submissions
            ADD CONSTRAINT fk_leaderboard_submissions_removed_by_admin
                FOREIGN KEY (removed_by_admin_id)
                    REFERENCES users (id)
                    ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_leaderboard_submissions_removed_at
    ON leaderboard_submissions (removed_at DESC);
