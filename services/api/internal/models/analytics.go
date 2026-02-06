package models

import "encoding/json"

type AnalyticsEventRequest struct {
	PlayToken string          `json:"play_token"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type AnalyticsEventResponse struct {
	Ok bool `json:"ok"`
}
