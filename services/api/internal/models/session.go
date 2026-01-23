package models

import "time"

type StartSessionRequest struct {
	GameID int64 `json:"game_id"`
}

type StartSessionResponse struct {
	PlayToken string    `json:"play_token"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}
