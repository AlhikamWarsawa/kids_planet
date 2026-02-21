package repos

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidPlayerID = errors.New("invalid player id")
)

type User struct {
	ID           int64
	PublicID     sql.NullString
	Name         string
	Email        string
	PasswordHash sql.NullString
	PinHash      sql.NullString
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
SELECT id, public_id, name, email, password_hash, pin_hash, role, status, created_at, updated_at
FROM users
WHERE email = $1
LIMIT 1;
`
	var u User
	err = r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.PublicID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
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
SELECT id, public_id, name, email, password_hash, pin_hash, role, status, created_at, updated_at
FROM users
WHERE id = $1
LIMIT 1;
`
	var u User
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID,
		&u.PublicID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
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

func (r *UserRepo) CreatePlayer(ctx context.Context, email string, pinHash string) (*User, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}

	pinHash = strings.TrimSpace(pinHash)
	if pinHash == "" {
		return nil, errors.New("pin hash is required")
	}

	publicID := uuid.NewString()
	name := playerNameFromEmail(email)

	const q = `
INSERT INTO users (public_id, name, email, pin_hash, role, status)
VALUES ($1::uuid, $2, $3, $4, 'player', 'active')
RETURNING id, public_id, name, email, password_hash, pin_hash, role, status, created_at, updated_at;
`

	var u User
	err = r.db.QueryRowContext(ctx, q, publicID, name, email, pinHash).Scan(
		&u.ID,
		&u.PublicID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("users.create_player: %w", err)
	}

	return &u, nil
}

func (r *UserRepo) FindPlayerByEmail(ctx context.Context, email string) (*User, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT id, public_id, name, email, password_hash, pin_hash, role, status, created_at, updated_at
FROM users
WHERE email = $1
  AND role = 'player'
LIMIT 1;
`

	var u User
	err = r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.PublicID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("users.find_player_by_email: %w", err)
	}
	return &u, nil
}

func (r *UserRepo) GetPlayerByID(ctx context.Context, playerID string) (*User, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, ErrInvalidPlayerID
	}
	if _, err := uuid.Parse(playerID); err != nil {
		return nil, ErrInvalidPlayerID
	}

	const q = `
SELECT id, public_id, name, email, password_hash, pin_hash, role, status, created_at, updated_at
FROM users
WHERE public_id = $1::uuid
  AND role = 'player'
LIMIT 1;
`

	var u User
	err := r.db.QueryRowContext(ctx, q, playerID).Scan(
		&u.ID,
		&u.PublicID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.PinHash,
		&u.Role,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("users.get_player_by_id: %w", err)
	}
	return &u, nil
}

func playerNameFromEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "player"
	}
	parts := strings.SplitN(email, "@", 2)
	name := strings.TrimSpace(parts[0])
	if name == "" {
		return "player"
	}
	return name
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
