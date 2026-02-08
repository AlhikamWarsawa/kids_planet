package repos

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, gameID int64, startedAt time.Time) (int64, error) {
	const q = `
INSERT INTO sessions (game_id, started_at)
VALUES ($1, $2)
RETURNING id;
`
	var id int64
	if err := r.db.QueryRowContext(ctx, q, gameID, startedAt).Scan(&id); err != nil {
		return 0, fmt.Errorf("sessions.create: %w", err)
	}
	return id, nil
}
