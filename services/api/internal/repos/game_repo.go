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

// ADMIN

type AdminGameFilter struct {
	Status sql.NullString
	Q      sql.NullString
	Page   int
	Limit  int
}

func (f *AdminGameFilter) normalize() {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Limit <= 0 {
		f.Limit = 24
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
}

func (r *GameRepo) ListAdminGames(ctx context.Context, filter AdminGameFilter) ([]Game, error) {
	filter.normalize()
	offset := (filter.Page - 1) * filter.Limit

	const q = `
SELECT id, title, slug, description, thumbnail, game_url, difficulty,
       age_category_id, free, status, created_by, created_at, updated_at
FROM games g
WHERE ($1::text IS NULL OR g.status = $1::game_status)
  AND ($2::text IS NULL OR g.title ILIKE '%'||$2||'%' OR g.slug ILIKE '%'||$2||'%')
ORDER BY g.updated_at DESC, g.id DESC
LIMIT $3 OFFSET $4;
`
	rows, err := r.db.QueryContext(ctx, q, filter.Status, filter.Q, filter.Limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Game, 0, filter.Limit)
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

func (r *GameRepo) CountAdminGames(ctx context.Context, filter AdminGameFilter) (int, error) {
	filter.normalize()

	const q = `
SELECT COUNT(*)
FROM games g
WHERE ($1::text IS NULL OR g.status = $1::game_status)
  AND ($2::text IS NULL OR g.title ILIKE '%'||$2||'%' OR g.slug ILIKE '%'||$2||'%');
`
	var total int
	if err := r.db.QueryRowContext(ctx, q, filter.Status, filter.Q).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *GameRepo) GetByIDAdmin(ctx context.Context, id int64) (*Game, error) {
	return r.GetByID(ctx, id)
}

func (r *GameRepo) SlugExists(ctx context.Context, slug string, excludeID *int64) (bool, error) {
	var ex sql.NullInt64
	if excludeID != nil && *excludeID > 0 {
		ex = sql.NullInt64{Int64: *excludeID, Valid: true}
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM games
  WHERE slug = $1
    AND ($2::bigint IS NULL OR id <> $2::bigint)
);
`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, slug, ex).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *GameRepo) AgeCategoryExists(ctx context.Context, id int64) (bool, error) {
	const q = `SELECT EXISTS (SELECT 1 FROM age_categories WHERE id = $1);`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

type CreateAdminGameInput struct {
	Title         string
	Slug          string
	Thumbnail     sql.NullString
	GameURL       sql.NullString
	AgeCategoryID int64
	Free          bool
	CreatedBy     int64
}

func (r *GameRepo) CreateGame(ctx context.Context, in CreateAdminGameInput) (int64, error) {
	const q = `
INSERT INTO games (title, slug, thumbnail, game_url, age_category_id, free, status, created_by)
VALUES ($1, $2, $3, $4, $5, $6, 'draft', $7)
RETURNING id;
`
	var id int64
	if err := r.db.QueryRowContext(ctx, q,
		in.Title,
		in.Slug,
		in.Thumbnail,
		in.GameURL,
		in.AgeCategoryID,
		in.Free,
		in.CreatedBy,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

type UpdateAdminGameInput struct {
	Title         *string
	Slug          *string
	Thumbnail     *string
	GameURL       *string
	AgeCategoryID *int64
	Free          *bool
}

func (r *GameRepo) UpdateGame(ctx context.Context, id int64, in UpdateAdminGameInput) (*Game, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const q = `
UPDATE games
SET
  title = COALESCE($2, title),
  slug = COALESCE($3, slug),
  thumbnail = COALESCE($4, thumbnail),
  game_url = COALESCE($5, game_url),
  age_category_id = COALESCE($6::bigint, age_category_id),
  free = COALESCE($7::boolean, free),
  updated_at = NOW()
WHERE id = $1
RETURNING id, title, slug, description, thumbnail, game_url, difficulty,
          age_category_id, free, status, created_by, created_at, updated_at;
`
	var g Game
	err := r.db.QueryRowContext(ctx, q,
		id,
		in.Title,
		in.Slug,
		in.Thumbnail,
		in.GameURL,
		in.AgeCategoryID,
		in.Free,
	).Scan(
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

func (r *GameRepo) SetStatus(ctx context.Context, id int64, status string) (*Game, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const q = `
UPDATE games
SET status = $2::game_status,
    updated_at = NOW()
WHERE id = $1
RETURNING id, title, slug, description, thumbnail, game_url, difficulty,
          age_category_id, free, status, created_by, created_at, updated_at;
`
	var g Game
	err := r.db.QueryRowContext(ctx, q, id, status).Scan(
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

func (r *GameRepo) SetGameURL(ctx context.Context, id int64, gameURL string) (*Game, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const q = `
UPDATE games
SET game_url = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, title, slug, description, thumbnail, game_url, difficulty,
          age_category_id, free, status, created_by, created_at, updated_at;
`
	var g Game
	err := r.db.QueryRowContext(ctx, q, id, gameURL).Scan(
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

func (r *GameRepo) ListAdmin(ctx context.Context, filter AdminGameFilter) ([]Game, error) {
	return r.ListAdminGames(ctx, filter)
}
func (r *GameRepo) CountAdmin(ctx context.Context, filter AdminGameFilter) (int, error) {
	return r.CountAdminGames(ctx, filter)
}

// PUBLIC

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

func (r *GameRepo) GetByIDPublic(ctx context.Context, id int64) (*GameListItem, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const q = `
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
WHERE g.id = $1
  AND g.status = 'active'
LIMIT 1;
`

	var it GameListItem
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&it.ID,
		&it.Title,
		&it.Slug,
		&it.Thumbnail,
		&it.GameURL,
		&it.AgeCategoryID,
		&it.Free,
		&it.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &it, nil
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

	const qNewest = `
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
ORDER BY g.created_at DESC
LIMIT $3 OFFSET $4;
`
	const qPopular = `
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
LEFT JOIN (
  SELECT ae.game_id, COUNT(*) AS popularity
  FROM analytics_events ae
  WHERE ae.event_name = 'game_start'
    AND ae.created_at >= NOW() - INTERVAL '7 days'
  GROUP BY ae.game_id
) pop ON pop.game_id = g.id
WHERE g.status = 'active'
  AND ($1::bigint IS NULL OR g.age_category_id = $1::bigint)
  AND ($2::bigint IS NULL OR EXISTS (
    SELECT 1
    FROM game_education_categories gec
    WHERE gec.game_id = g.id
      AND gec.education_category_id = $2::bigint
  ))
ORDER BY COALESCE(pop.popularity, 0) DESC, g.created_at DESC, g.id DESC
LIMIT $3 OFFSET $4;
`

	query := qNewest
	if filter.Sort == GameSortPopular {
		query = qPopular
	}

	rows, err := r.db.QueryContext(ctx, query,
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

func (r *GameRepo) CreateAdminGame(ctx context.Context, in CreateAdminGameInput) (*Game, error) {
	id, err := r.CreateGame(ctx, in)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *GameRepo) UpdateAdminGame(ctx context.Context, id int64, in UpdateAdminGameInput) (*Game, error) {
	return r.UpdateGame(ctx, id, in)
}
