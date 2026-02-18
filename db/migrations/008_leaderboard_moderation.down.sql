ALTER TABLE leaderboard_submissions
    DROP CONSTRAINT IF EXISTS fk_leaderboard_submissions_removed_by_admin;

DROP INDEX IF EXISTS idx_leaderboard_submissions_removed_at;

ALTER TABLE leaderboard_submissions
    DROP COLUMN IF EXISTS removed_by_admin_id,
    DROP COLUMN IF EXISTS removed_at,
    DROP COLUMN IF EXISTS member;
