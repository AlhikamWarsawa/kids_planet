package services

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type GameService struct {
	gameRepo *repos.GameRepo
}

func NewGameService(gameRepo *repos.GameRepo) *GameService {
	return &GameService{gameRepo: gameRepo}
}

type ListPublicGamesInput struct {
	AgeCategoryID       *int64
	EducationCategoryID *int64
	Sort                string
	Page                int
	Limit               int
}

type GameListDTO struct {
	Items []GameListItemDTO `json:"items"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int               `json:"total"`
}

type GameListItemDTO struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Thumbnail     *string `json:"thumbnail"`
	GameURL       *string `json:"game_url"`
	AgeCategoryID int64   `json:"age_category_id"`
	Free          bool    `json:"free"`
	CreatedAt     string  `json:"created_at"`
}

type GameDetailDTO struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Thumbnail     *string `json:"thumbnail"`
	GameURL       *string `json:"game_url"`
	AgeCategoryID int64   `json:"age_category_id"`
	Free          bool    `json:"free"`
	CreatedAt     string  `json:"created_at"`
}

func (s *GameService) ListPublicGames(ctx context.Context, in ListPublicGamesInput) (*GameListDTO, error) {
	page := in.Page
	limit := in.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 24
	}

	if page < 1 {
		return nil, utils.ErrBadRequest("page must be >= 1")
	}
	if limit < 1 || limit > 100 {
		return nil, utils.ErrBadRequest("limit must be between 1 and 100")
	}

	sort := strings.TrimSpace(strings.ToLower(in.Sort))
	if sort == "" {
		sort = "newest"
	}
	var sortEnum repos.GameListSort
	switch sort {
	case "newest":
		sortEnum = repos.GameSortNewest
	case "popular":
		sortEnum = repos.GameSortPopular
	default:
		return nil, utils.ErrBadRequest("sort must be one of: newest, popular")
	}

	var age sql.NullInt64
	if in.AgeCategoryID != nil {
		if *in.AgeCategoryID < 1 {
			return nil, utils.ErrBadRequest("age_category_id must be >= 1")
		}
		age = sql.NullInt64{Int64: *in.AgeCategoryID, Valid: true}
	}

	var edu sql.NullInt64
	if in.EducationCategoryID != nil {
		if *in.EducationCategoryID < 1 {
			return nil, utils.ErrBadRequest("education_category_id must be >= 1")
		}
		edu = sql.NullInt64{Int64: *in.EducationCategoryID, Valid: true}
	}

	filter := repos.GameListFilter{
		AgeCategoryID:       age,
		EducationCategoryID: edu,
		Sort:                sortEnum,
		Page:                page,
		Limit:               limit,
	}

	items, err := s.gameRepo.ListPublic(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	total, err := s.gameRepo.CountPublic(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	out := make([]GameListItemDTO, 0, len(items))
	for _, it := range items {
		var thumb *string
		if it.Thumbnail.Valid {
			v := it.Thumbnail.String
			thumb = &v
		}

		var url *string
		if it.GameURL.Valid {
			v := it.GameURL.String
			url = &v
		}

		out = append(out, GameListItemDTO{
			ID:            it.ID,
			Title:         it.Title,
			Slug:          it.Slug,
			Thumbnail:     thumb,
			GameURL:       url,
			AgeCategoryID: it.AgeCategoryID,
			Free:          it.Free,
			CreatedAt:     it.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	return &GameListDTO{
		Items: out,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *GameService) GetPublicGameByID(ctx context.Context, id int64) (*GameDetailDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	it, err := s.gameRepo.GetByIDPublic(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	var thumb *string
	if it.Thumbnail.Valid {
		v := it.Thumbnail.String
		thumb = &v
	}

	var url *string
	if it.GameURL.Valid {
		v := it.GameURL.String
		url = &v
	}

	return &GameDetailDTO{
		ID:            it.ID,
		Title:         it.Title,
		Slug:          it.Slug,
		Thumbnail:     thumb,
		GameURL:       url,
		AgeCategoryID: it.AgeCategoryID,
		Free:          it.Free,
		CreatedAt:     it.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}, nil
}
