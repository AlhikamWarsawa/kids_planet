package models

type SubmitScoreRequest struct {
	GameID int64 `json:"game_id"`
	Score  int   `json:"score"`
}

type SubmitScoreResponse struct {
	Accepted  bool `json:"accepted"`
	BestScore int  `json:"best_score"`
}
