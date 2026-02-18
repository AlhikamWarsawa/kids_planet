package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type LeaderboardSubmission struct {
	ID               int64
	GameID           int64
	PlayerID         sql.NullInt64
	SessionID        sql.NullString
	Member           sql.NullString
	Score            int
	IPHash           sql.NullString
	UserAgentHash    sql.NullString
	Flagged          bool
	FlagReason       sql.NullString
	CreatedAt        time.Time
	UpdatedAt        time.Time
	RemovedByAdminID sql.NullInt64
	RemovedAt        sql.NullTime
}

type SubmissionRepo struct {
	db *sql.DB
}

func NewSubmissionRepo(db *sql.DB) *SubmissionRepo {
	return &SubmissionRepo{db: db}
}

func (r *SubmissionRepo) CreateSubmission(ctx context.Context, s *LeaderboardSubmission) (int64, error) {
	if s == nil {
		return 0, errors.New("submission is required")
	}

	const q = `
INSERT INTO leaderboard_submissions
  (game_id, player_id, session_id, member, score, ip_hash, user_agent_hash, flagged, flag_reason)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at;
`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, q,
		s.GameID,
		s.PlayerID,
		s.SessionID,
		s.Member,
		s.Score,
		s.IPHash,
		s.UserAgentHash,
		s.Flagged,
		s.FlagReason,
	).Scan(&id, &createdAt)
	if err != nil {
		return 0, fmt.Errorf("leaderboard_submissions.create: %w", err)
	}

	s.ID = id
	s.CreatedAt = createdAt

	return id, nil
}

type FlaggedSubmission struct {
	ID         int64
	GameID     int64
	PlayerName string
	Score      int
	Flagged    bool
	FlagReason sql.NullString
	SessionID  sql.NullString
	CreatedAt  time.Time
}

func (r *SubmissionRepo) ListFlagged(ctx context.Context, limit int) ([]FlaggedSubmission, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	const q = `
SELECT
  s.id,
  s.game_id,
  COALESCE(p.nickname, '') AS player_name,
  s.score,
  s.flagged,
  s.flag_reason,
  s.session_id,
  s.created_at
FROM leaderboard_submissions s
LEFT JOIN players p ON p.id = s.player_id
WHERE s.flagged = TRUE
  AND s.removed_at IS NULL
ORDER BY s.created_at DESC
LIMIT $1;
`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]FlaggedSubmission, 0, limit)
	for rows.Next() {
		var it FlaggedSubmission
		if err := rows.Scan(
			&it.ID,
			&it.GameID,
			&it.PlayerName,
			&it.Score,
			&it.Flagged,
			&it.FlagReason,
			&it.SessionID,
			&it.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SubmissionRepo) GetByID(ctx context.Context, id int64) (*LeaderboardSubmission, error) {
	if id <= 0 {
		return nil, errors.New("submission id is required")
	}

	const q = `
SELECT
  s.id,
  s.game_id,
  s.player_id,
  s.session_id,
  s.member,
  s.score,
  s.ip_hash,
  s.user_agent_hash,
  s.flagged,
  s.flag_reason,
  s.created_at,
  s.updated_at,
  s.removed_by_admin_id,
  s.removed_at
FROM leaderboard_submissions s
WHERE s.id = $1
LIMIT 1;
`
	var it LeaderboardSubmission
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&it.ID,
		&it.GameID,
		&it.PlayerID,
		&it.SessionID,
		&it.Member,
		&it.Score,
		&it.IPHash,
		&it.UserAgentHash,
		&it.Flagged,
		&it.FlagReason,
		&it.CreatedAt,
		&it.UpdatedAt,
		&it.RemovedByAdminID,
		&it.RemovedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("leaderboard_submissions.get_by_id: %w", err)
	}
	return &it, nil
}

func (r *SubmissionRepo) MarkRemovedByAdmin(ctx context.Context, id int64, adminID int64, now time.Time) error {
	if id <= 0 {
		return errors.New("submission id is required")
	}
	if adminID <= 0 {
		return errors.New("admin id is required")
	}

	if now.IsZero() {
		now = time.Now().UTC()
	}

	const q = `
UPDATE leaderboard_submissions
SET
  flagged = TRUE,
  flag_reason = 'removed_by_admin',
  updated_at = $1,
  removed_by_admin_id = $2,
  removed_at = $1
WHERE id = $3;
`
	res, err := r.db.ExecContext(ctx, q, now, adminID, id)
	if err != nil {
		return fmt.Errorf("leaderboard_submissions.mark_removed: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("leaderboard_submissions.mark_removed: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
