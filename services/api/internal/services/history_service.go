package services

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type HistoryService struct {
	repo *repos.PlayerHistoryRepo
}

func NewHistoryService(repo *repos.PlayerHistoryRepo) *HistoryService {
	return &HistoryService{repo: repo}
}

func (s *HistoryService) ListPlayerHistory(
	ctx context.Context,
	playerID string,
	page int,
	limit int,
) (*models.PlayerHistoryResponse, error) {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return nil, utils.ErrUnauthorized()
	}
	if _, err := uuid.Parse(playerID); err != nil {
		return nil, utils.ErrUnauthorized()
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, utils.ErrBadRequest("limit must be an integer between 1 and 100")
	}

	items, total, err := s.repo.ListByPlayerID(ctx, playerID, page, limit)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	out := make([]models.PlayerHistoryItem, 0, len(items))
	for _, it := range items {
		var scorePtr *int
		var statusPtr *string

		if it.Score.Valid {
			score := int(it.Score.Int64)
			scorePtr = &score
			status := "completed"
			statusPtr = &status
		}

		out = append(out, models.PlayerHistoryItem{
			GameID:   it.GameID,
			Title:    it.Title,
			PlayedAt: it.PlayedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Score:    scorePtr,
			Status:   statusPtr,
		})
	}

	return &models.PlayerHistoryResponse{
		Data: out,
		Pagination: models.PlayerHistoryPagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}, nil
}
