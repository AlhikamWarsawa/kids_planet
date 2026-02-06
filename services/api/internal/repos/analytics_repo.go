package repos

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type AnalyticsRepo struct {
	db *sql.DB
}

func NewAnalyticsRepo(db *sql.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) InsertAnalyticsEvent(
	ctx context.Context,
	sessionID string,
	gameID int64,
	eventName string,
	eventData *string,
	ip string,
	userAgent string,
) error {
	const q = `
INSERT INTO analytics_events
  (session_id, game_id, event_name, event_data, ip, user_agent)
VALUES
  ($1, $2, $3, $4, $5, $6);
`

	dataVal := nullString(eventData)
	ipVal := nullStringPtr(ip)
	uaVal := nullStringPtr(userAgent)

	if _, err := r.db.ExecContext(ctx, q,
		sessionID,
		gameID,
		eventName,
		dataVal,
		ipVal,
		uaVal,
	); err != nil {
		return fmt.Errorf("analytics_events.insert: %w", err)
	}

	return nil
}

func nullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{Valid: false}
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: trimmed, Valid: true}
}

func nullStringPtr(v string) sql.NullString {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: trimmed, Valid: true}
}
