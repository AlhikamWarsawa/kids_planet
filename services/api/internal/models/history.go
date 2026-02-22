package models

type PlayerHistoryItem struct {
	GameID   int64   `json:"game_id"`
	Title    string  `json:"title"`
	PlayedAt string  `json:"played_at"`
	Score    *int    `json:"score,omitempty"`
	Status   *string `json:"status,omitempty"`
}

type PlayerHistoryPagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type PlayerHistoryResponse struct {
	Data       []PlayerHistoryItem     `json:"data"`
	Pagination PlayerHistoryPagination `json:"pagination"`
}
