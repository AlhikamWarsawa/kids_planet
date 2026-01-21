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

type GameListSort string

const (
	GameSortNewest  GameListSort = "newest"
	GameSortPopular GameListSort = "popular"
)

type GameListFilter struct {
	AgeCategoryID       sql.NullInt64
	EducationCategoryID sql.NullInt64
	Sort                GameListSort
	Page                int
	Limit               int
}

type GameListItem struct {
	ID            int64
	Title         string
	Slug          string
	Thumbnail     sql.NullString
	GameURL       sql.NullString
	AgeCategoryID int64
	Free          bool
	CreatedAt     time.Time
}

func (f *GameListFilter) normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 24
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	switch f.Sort {
	case GameSortNewest, GameSortPopular:
	default:
		f.Sort = GameSortNewest
	}
}

func (r *GameRepo) ListPublic(ctx context.Context, filter GameListFilter) ([]GameListItem, error) {
	filter.normalize()

	offset := (filter.Page - 1) * filter.Limit

	orderBy := "g.created_at DESC"
	switch filter.Sort {
	case GameSortPopular:
		orderBy = "g.created_at DESC"
	case GameSortNewest:
		orderBy = "g.created_at DESC"
	}

	q := `
SELECT
  g.id,
  g.title,
  g.slug,
  g.thumbnail,
  g.game_url,
  g.age_category_id,
  g.free,
  g.created_at
FROM games g
WHERE g.status = 'active'
  AND ($1::bigint IS NULL OR g.age_category_id = $1::bigint)
  AND ($2::bigint IS NULL OR EXISTS (
    SELECT 1
    FROM game_education_categories gec
    WHERE gec.game_id = g.id
      AND gec.education_category_id = $2::bigint
  ))
ORDER BY ` + orderBy + `
LIMIT $3 OFFSET $4;
`

	rows, err := r.db.QueryContext(ctx, q,
		filter.AgeCategoryID,
		filter.EducationCategoryID,
		filter.Limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]GameListItem, 0, filter.Limit)
	for rows.Next() {
		var it GameListItem
		if err := rows.Scan(
			&it.ID,
			&it.Title,
			&it.Slug,
			&it.Thumbnail,
			&it.GameURL,
			&it.AgeCategoryID,
			&it.Free,
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

func (r *GameRepo) CountPublic(ctx context.Context, filter GameListFilter) (int, error) {
	filter.normalize()

	const q = `
SELECT COUNT(*)
FROM games g
WHERE g.status = 'active'
  AND ($1::bigint IS NULL OR g.age_category_id = $1::bigint)
  AND ($2::bigint IS NULL OR EXISTS (
    SELECT 1
    FROM game_education_categories gec
    WHERE gec.game_id = g.id
      AND gec.education_category_id = $2::bigint
  ));
`

	var total int
	if err := r.db.QueryRowContext(ctx, q,
		filter.AgeCategoryID,
		filter.EducationCategoryID,
	).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
