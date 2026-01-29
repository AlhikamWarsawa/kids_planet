package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AgeCategory struct {
	ID        int64
	Label     string
	MinAge    int
	MaxAge    int
	CreatedAt time.Time
}

type AgeCategoryRepo struct {
	db *sql.DB
}

func NewAgeCategoryRepo(db *sql.DB) *AgeCategoryRepo {
	return &AgeCategoryRepo{db: db}
}

func (r *AgeCategoryRepo) LabelExists(ctx context.Context, label string, excludeID *int64) (bool, error) {
	var ex sql.NullInt64
	if excludeID != nil && *excludeID > 0 {
		ex = sql.NullInt64{Int64: *excludeID, Valid: true}
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM age_categories
  WHERE label = $1
    AND ($2::bigint IS NULL OR id <> $2::bigint)
);
`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, label, ex).Scan(&exists); err != nil {
		return false, fmt.Errorf("age_categories.label_exists: %w", err)
	}
	return exists, nil
}

type AgeCategoryListFilter struct {
	Q     sql.NullString
	Page  int
	Limit int
}

func (f *AgeCategoryListFilter) normalize() {
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

func (r *AgeCategoryRepo) List(ctx context.Context, q string, page, limit int) ([]AgeCategory, error) {
	filter := AgeCategoryListFilter{
		Page:  page,
		Limit: limit,
	}
	if q != "" {
		filter.Q = sql.NullString{String: q, Valid: true}
	}
	filter.normalize()

	offset := (filter.Page - 1) * filter.Limit

	const query = `
SELECT id, label, min_age, max_age, created_at
FROM age_categories ac
WHERE ($1::text IS NULL OR ac.label ILIKE '%'||$1||'%')
ORDER BY ac.min_age ASC, ac.max_age ASC, ac.id ASC
LIMIT $2 OFFSET $3;
`
	rows, err := r.db.QueryContext(ctx, query, filter.Q, filter.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("age_categories.list: %w", err)
	}
	defer rows.Close()

	out := make([]AgeCategory, 0, filter.Limit)
	for rows.Next() {
		var it AgeCategory
		if err := rows.Scan(
			&it.ID,
			&it.Label,
			&it.MinAge,
			&it.MaxAge,
			&it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("age_categories.list.scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("age_categories.list.rows: %w", err)
	}
	return out, nil
}

type CreateAgeCategoryInput struct {
	Label  string
	MinAge int
	MaxAge int
}

func (r *AgeCategoryRepo) Create(ctx context.Context, in CreateAgeCategoryInput) (*AgeCategory, error) {
	const query = `
INSERT INTO age_categories (label, min_age, max_age)
VALUES ($1, $2, $3)
RETURNING id, label, min_age, max_age, created_at;
`
	var it AgeCategory
	err := r.db.QueryRowContext(ctx, query,
		in.Label,
		in.MinAge,
		in.MaxAge,
	).Scan(
		&it.ID,
		&it.Label,
		&it.MinAge,
		&it.MaxAge,
		&it.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("age_categories.create: %w", err)
	}
	return &it, nil
}

type UpdateAgeCategoryInput struct {
	Label  *string
	MinAge *int
	MaxAge *int
}

func (r *AgeCategoryRepo) Update(ctx context.Context, id int64, in UpdateAgeCategoryInput) (*AgeCategory, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const query = `
UPDATE age_categories
SET
  label  = COALESCE($2, label),
  min_age = COALESCE($3::int, min_age),
  max_age = COALESCE($4::int, max_age)
WHERE id = $1
RETURNING id, label, min_age, max_age, created_at;
`
	var it AgeCategory
	err := r.db.QueryRowContext(ctx, query,
		id,
		in.Label,
		in.MinAge,
		in.MaxAge,
	).Scan(
		&it.ID,
		&it.Label,
		&it.MinAge,
		&it.MaxAge,
		&it.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("age_categories.update: %w", err)
	}
	return &it, nil
}

func (r *AgeCategoryRepo) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id is required")
	}

	const query = `DELETE FROM age_categories WHERE id = $1;`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("age_categories.delete: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("age_categories.delete.rows_affected: %w", err)
	}
	if ra == 0 {
		return ErrNotFound
	}
	return nil
}
