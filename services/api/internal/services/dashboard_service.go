package services

import (
	"context"

	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type DashboardService struct {
	dashboardRepo *repos.DashboardRepo
}

func NewDashboardService(dashboardRepo *repos.DashboardRepo) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

type DashboardTopGameDTO struct {
	GameID int64  `json:"game_id"`
	Title  string `json:"title"`
	Plays  int    `json:"plays"`
}

type DashboardOverviewDTO struct {
	SessionsToday    int                   `json:"sessions_today"`
	TopGames         []DashboardTopGameDTO `json:"top_games"`
	TotalActiveGames int                   `json:"total_active_games"`
	TotalPlayers     int                   `json:"total_players"`
}

func (s *DashboardService) GetOverview(ctx context.Context) (*DashboardOverviewDTO, *utils.AppError) {
	sessionsToday, err := s.dashboardRepo.CountSessionsTodayUTC(ctx)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}

	topGames, err := s.dashboardRepo.ListTopGames(ctx, 5)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}

	activeGames, err := s.dashboardRepo.CountActiveGames(ctx)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}

	totalPlayers, err := s.dashboardRepo.CountPlayers(ctx)
	if err != nil {
		ae := utils.ErrInternal()
		return nil, &ae
	}

	outTop := make([]DashboardTopGameDTO, 0, len(topGames))
	for _, it := range topGames {
		outTop = append(outTop, DashboardTopGameDTO{
			GameID: it.GameID,
			Title:  it.Title,
			Plays:  it.Plays,
		})
	}

	return &DashboardOverviewDTO{
		SessionsToday:    sessionsToday,
		TopGames:         outTop,
		TotalActiveGames: activeGames,
		TotalPlayers:     totalPlayers,
	}, nil
}
