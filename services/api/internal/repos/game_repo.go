package repos

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Game struct {
	ID            int64
	Title         string
	Slug          string
	Description   sql.NullString
	Thumbnail     sql.NullString
	GameURL       sql.NullString
	Difficulty    sql.NullString
	AgeCategoryID int64
	Free          bool
	Status        string
	CreatedBy     int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type GameRepo struct {
	db *sql.DB
}

func NewGameRepo(db *sql.DB) *GameRepo {
	return &GameRepo{db: db}
}

func (r *GameRepo) ListBasic(ctx context.Context, limit int) ([]Game, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	const q = `
SELECT id, title, slug, description, thumbnail, game_url, difficulty,
       age_category_id, free, status, created_by, created_at, updated_at
FROM games
ORDER BY created_at DESC
LIMIT $1;
`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Game, 0, limit)
	for rows.Next() {
		var g Game
		if err := rows.Scan(
			&g.ID,
			&g.Title,
			&g.Slug,
			&g.Description,
			&g.Thumbnail,
			&g.GameURL,
			&g.Difficulty,
			&g.AgeCategoryID,
			&g.Free,
			&g.Status,
			&g.CreatedBy,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *GameRepo) GetByID(ctx context.Context, id int64) (*Game, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const q = `
SELECT id, title, slug, description, thumbnail, game_url, difficulty,
       age_category_id, free, status, created_by, created_at, updated_at
FROM games
WHERE id = $1
LIMIT 1;
`
	var g Game
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&g.ID,
		&g.Title,
		&g.Slug,
		&g.Description,
		&g.Thumbnail,
		&g.GameURL,
		&g.Difficulty,
		&g.AgeCategoryID,
		&g.Free,
		&g.Status,
		&g.CreatedBy,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}
