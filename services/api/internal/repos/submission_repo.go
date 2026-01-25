package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type LeaderboardSubmission struct {
	ID            int64
	GameID        int64
	PlayerID      sql.NullInt64
	SessionID     sql.NullString
	Score         int
	IPHash        sql.NullString
	UserAgentHash sql.NullString
	Flagged       bool
	FlagReason    sql.NullString
	CreatedAt     time.Time
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
  (game_id, player_id, session_id, score, ip_hash, user_agent_hash, flagged, flag_reason)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at;
`

	var id int64
	var createdAt time.Time
	err := r.db.QueryRowContext(ctx, q,
		s.GameID,
		s.PlayerID,
		s.SessionID,
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
