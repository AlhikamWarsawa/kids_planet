package repos

import (
	"database/sql"
)

type SubmissionRepo struct {
	db *sql.DB
}

func NewSubmissionRepo(db *sql.DB) *SubmissionRepo {
	return &SubmissionRepo{db: db}
}
