package models

type LeaderboardItem struct {
	Member string `json:"member"`
	Score  int    `json:"score"`
}

type LeaderboardViewResponse struct {
	GameID int64             `json:"game_id"`
	Period string            `json:"period"`
	Scope  string            `json:"scope"`
	Limit  int               `json:"limit"`
	Items  []LeaderboardItem `json:"items"`
}

type LeaderboardSelfDTO struct {
	GameID int64  `json:"game_id"`
	Rank   *int64 `json:"rank,omitempty"`
	Score  *int64 `json:"score,omitempty"`
	Period string `json:"period"`
	Scope  string `json:"scope"`
}
