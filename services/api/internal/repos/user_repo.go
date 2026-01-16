package repos

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("not found")

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

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return nil, errors.New("email is required")
	}

	const q = `
SELECT id, name, email, password_hash, role, status, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;
`
	var u User
	err := r.db.QueryRowContext(ctx, q, email).Scan(
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
		return nil, err
	}
	return &u, nil
}
