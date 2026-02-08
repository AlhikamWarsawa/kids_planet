package repos

import (
	"context"
	"database/sql"
	"fmt"
)

type DashboardRepo struct {
	db *sql.DB
}

func NewDashboardRepo(db *sql.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

type TopGameRow struct {
	GameID int64
	Title  string
	Plays  int
}

func (r *DashboardRepo) CountSessionsTodayUTC(ctx context.Context) (int, error) {
	const q = `
SELECT COUNT(*)
FROM sessions
WHERE started_at >= date_trunc('day', now() AT TIME ZONE 'utc');
`
	var total int
	if err := r.db.QueryRowContext(ctx, q).Scan(&total); err != nil {
		return 0, fmt.Errorf("dashboard.sessions_today: %w", err)
	}
	return total, nil
}

func (r *DashboardRepo) ListTopGames(ctx context.Context, limit int) ([]TopGameRow, error) {
	if limit <= 0 {
		limit = 5
	}

	const q = `
SELECT s.game_id, g.title, COUNT(*) AS plays
FROM sessions s
JOIN games g ON g.id = s.game_id
WHERE g.status = 'active'
GROUP BY s.game_id, g.title
ORDER BY plays DESC, s.game_id ASC
LIMIT $1;
`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("dashboard.top_games: %w", err)
	}
	defer rows.Close()

	out := make([]TopGameRow, 0, limit)
	for rows.Next() {
		var it TopGameRow
		if err := rows.Scan(&it.GameID, &it.Title, &it.Plays); err != nil {
			return nil, fmt.Errorf("dashboard.top_games.scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("dashboard.top_games.rows: %w", err)
	}
	return out, nil
}

func (r *DashboardRepo) CountActiveGames(ctx context.Context) (int, error) {
	const q = `
SELECT COUNT(*)
FROM games
WHERE status = 'active';
`
	var total int
	if err := r.db.QueryRowContext(ctx, q).Scan(&total); err != nil {
		return 0, fmt.Errorf("dashboard.active_games: %w", err)
	}
	return total, nil
}

func (r *DashboardRepo) CountPlayers(ctx context.Context) (int, error) {
	const q = `
SELECT COUNT(*)
FROM players;
`
	var total int
	if err := r.db.QueryRowContext(ctx, q).Scan(&total); err != nil {
		return 0, fmt.Errorf("dashboard.total_players: %w", err)
	}
	return total, nil
}
