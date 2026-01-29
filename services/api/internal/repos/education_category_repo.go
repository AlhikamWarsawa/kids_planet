package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type EducationCategory struct {
	ID        int64
	Name      string
	Icon      sql.NullString
	Color     sql.NullString
	CreatedAt time.Time
}

type EducationCategoryRepo struct {
	db *sql.DB
}

func NewEducationCategoryRepo(db *sql.DB) *EducationCategoryRepo {
	return &EducationCategoryRepo{db: db}
}

type EducationCategoryListFilter struct {
	Q     sql.NullString
	Page  int
	Limit int
}

func (f *EducationCategoryListFilter) normalize() {
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

func (r *EducationCategoryRepo) List(ctx context.Context, q string, page, limit int) ([]EducationCategory, error) {
	filter := EducationCategoryListFilter{
		Page:  page,
		Limit: limit,
	}
	if q != "" {
		filter.Q = sql.NullString{String: q, Valid: true}
	}
	filter.normalize()

	offset := (filter.Page - 1) * filter.Limit

	const query = `
SELECT id, name, icon, color, created_at
FROM education_categories ec
WHERE ($1::text IS NULL OR ec.name ILIKE '%'||$1||'%')
ORDER BY ec.name ASC, ec.id ASC
LIMIT $2 OFFSET $3;
`
	rows, err := r.db.QueryContext(ctx, query, filter.Q, filter.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("education_categories.list: %w", err)
	}
	defer rows.Close()

	out := make([]EducationCategory, 0, filter.Limit)
	for rows.Next() {
		var it EducationCategory
		if err := rows.Scan(
			&it.ID,
			&it.Name,
			&it.Icon,
			&it.Color,
			&it.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("education_categories.list.scan: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("education_categories.list.rows: %w", err)
	}
	return out, nil
}

type CreateEducationCategoryInput struct {
	Name  string
	Icon  sql.NullString
	Color sql.NullString
}

func (r *EducationCategoryRepo) Create(ctx context.Context, in CreateEducationCategoryInput) (*EducationCategory, error) {
	const query = `
INSERT INTO education_categories (name, icon, color)
VALUES ($1, $2, $3)
RETURNING id, name, icon, color, created_at;
`
	var it EducationCategory
	err := r.db.QueryRowContext(ctx, query,
		in.Name,
		in.Icon,
		in.Color,
	).Scan(
		&it.ID,
		&it.Name,
		&it.Icon,
		&it.Color,
		&it.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("education_categories.create: %w", err)
	}
	return &it, nil
}

type UpdateEducationCategoryInput struct {
	Name  *string
	Icon  *string
	Color *string
}

func (r *EducationCategoryRepo) Update(ctx context.Context, id int64, in UpdateEducationCategoryInput) (*EducationCategory, error) {
	if id <= 0 {
		return nil, errors.New("id is required")
	}

	const query = `
UPDATE education_categories
SET
  name = COALESCE($2, name),
  icon = COALESCE($3, icon),
  color = COALESCE($4, color)
WHERE id = $1
RETURNING id, name, icon, color, created_at;
`
	var it EducationCategory
	err := r.db.QueryRowContext(ctx, query,
		id,
		in.Name,
		in.Icon,
		in.Color,
	).Scan(
		&it.ID,
		&it.Name,
		&it.Icon,
		&it.Color,
		&it.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("education_categories.update: %w", err)
	}
	return &it, nil
}

func (r *EducationCategoryRepo) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("id is required")
	}

	const query = `DELETE FROM education_categories WHERE id = $1;`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("education_categories.delete: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("education_categories.delete.rows_affected: %w", err)
	}
	if ra == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *EducationCategoryRepo) NameExists(ctx context.Context, name string, excludeID *int64) (bool, error) {
	var ex sql.NullInt64
	if excludeID != nil && *excludeID > 0 {
		ex = sql.NullInt64{Int64: *excludeID, Valid: true}
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM education_categories
  WHERE name = $1
    AND ($2::bigint IS NULL OR id <> $2::bigint)
);
`
	var exists bool
	if err := r.db.QueryRowContext(ctx, q, name, ex).Scan(&exists); err != nil {
		return false, fmt.Errorf("education_categories.name_exists: %w", err)
	}
	return exists, nil
}
