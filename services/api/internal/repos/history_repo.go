package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PlayerHistoryRow struct {
	GameID   int64
	Title    string
	PlayedAt time.Time
	Score    sql.NullInt64
}

type PlayerHistoryRepo struct {
	db *sql.DB
}

func NewPlayerHistoryRepo(db *sql.DB) *PlayerHistoryRepo {
	return &PlayerHistoryRepo{db: db}
}

func (r *PlayerHistoryRepo) ListByPlayerID(
	ctx context.Context,
	playerID string,
	page int,
	limit int,
) ([]PlayerHistoryRow, int, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, 0, ErrInvalidPlayerID
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	const countQ = `
WITH player_events AS (
  SELECT
    ae.session_id,
    ae.game_id,
    MAX(ae.created_at) AS played_at
  FROM analytics_events ae
  WHERE COALESCE(ae.event_data->>'player_id', '') = $1
    AND (ae.event_name = 'game_start' OR ae.event_name LIKE 'gameplay%')
  GROUP BY ae.session_id, ae.game_id
)
SELECT COUNT(*) FROM player_events;
`

	var total int
	if err := r.db.QueryRowContext(ctx, countQ, playerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("player_history.count: %w", err)
	}

	if total == 0 {
		return []PlayerHistoryRow{}, 0, nil
	}

	const listQ = `
WITH player_events AS (
  SELECT
    ae.session_id,
    ae.game_id,
    MAX(ae.created_at) AS played_at
  FROM analytics_events ae
  WHERE COALESCE(ae.event_data->>'player_id', '') = $1
    AND (ae.event_name = 'game_start' OR ae.event_name LIKE 'gameplay%')
  GROUP BY ae.session_id, ae.game_id
),
scores AS (
  SELECT
    ls.session_id,
    ls.game_id,
    MAX(ls.score)::bigint AS best_score
  FROM leaderboard_submissions ls
  WHERE ls.session_id IS NOT NULL
  GROUP BY ls.session_id, ls.game_id
)
SELECT
  pe.game_id,
  g.title,
  pe.played_at,
  sc.best_score
FROM player_events pe
JOIN games g ON g.id = pe.game_id
LEFT JOIN scores sc
  ON sc.session_id = pe.session_id
 AND sc.game_id = pe.game_id
ORDER BY pe.played_at DESC, pe.game_id DESC
LIMIT $2 OFFSET $3;
`

	rows, err := r.db.QueryContext(ctx, listQ, playerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("player_history.list: %w", err)
	}
	defer rows.Close()

	out := make([]PlayerHistoryRow, 0, limit)
	for rows.Next() {
		var row PlayerHistoryRow
		if err := rows.Scan(
			&row.GameID,
			&row.Title,
			&row.PlayedAt,
			&row.Score,
		); err != nil {
			return nil, 0, fmt.Errorf("player_history.list.scan: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("player_history.list.rows: %w", err)
	}

	return out, total, nil
}
