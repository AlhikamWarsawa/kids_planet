package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidEmail = errors.New("invalid email")
)

type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash sql.NullString
	Role         string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func NormalizeEmail(email string) (string, error) {
	e := strings.TrimSpace(strings.ToLower(email))
	if e == "" {
		return "", ErrInvalidEmail
	}
	if !strings.Contains(e, "@") {
		return "", ErrInvalidEmail
	}
	return e, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT id, name, email, password_hash, role, status, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;
`
	var u User
	err = r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("users.find_by_email: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, ErrNotFound
	}

	const q = `
SELECT id, name, email, password_hash, role, status, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;
`
	var u User
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("users.find_by_id: %w", err)
	}
	return &u, nil
}
