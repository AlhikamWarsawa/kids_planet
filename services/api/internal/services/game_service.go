package services

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
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

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type AdminListGamesInput struct {
	Status string
	Q      string
	Page   int
	Limit  int
}

type AdminGameListDTO struct {
	Items []models.AdminGameDTO `json:"items"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
	Total int                   `json:"total"`
}

func (s *GameService) ListAdminGames(ctx context.Context, in AdminListGamesInput) (*AdminGameListDTO, error) {
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

	status := strings.TrimSpace(strings.ToLower(in.Status))
	var statusNS sql.NullString
	if status != "" {
		if status != "draft" && status != "active" && status != "archived" {
			return nil, utils.ErrBadRequest("status must be one of: draft, active, archived")
		}
		statusNS = sql.NullString{String: status, Valid: true}
	}

	q := strings.TrimSpace(in.Q)
	var qNS sql.NullString
	if q != "" {
		qNS = sql.NullString{String: q, Valid: true}
	}

	filter := repos.AdminGameFilter{
		Status: statusNS,
		Q:      qNS,
		Page:   page,
		Limit:  limit,
	}

	items, err := s.gameRepo.ListAdmin(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	total, err := s.gameRepo.CountAdmin(ctx, filter)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	out := make([]models.AdminGameDTO, 0, len(items))
	for _, g := range items {
		out = append(out, toAdminGameDTO(g))
	}

	return &AdminGameListDTO{
		Items: out,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *GameService) CreateAdminGame(ctx context.Context, createdBy int64, req models.CreateGameRequest) (*models.AdminGameDTO, error) {
	title := strings.TrimSpace(req.Title)
	slug := strings.TrimSpace(req.Slug)

	if title == "" {
		return nil, utils.ErrBadRequest("title is required")
	}
	if len(title) > 150 {
		return nil, utils.ErrBadRequest("title must be <= 150 chars")
	}
	if slug == "" {
		return nil, utils.ErrBadRequest("slug is required")
	}
	if len(slug) > 150 {
		return nil, utils.ErrBadRequest("slug must be <= 150 chars")
	}
	if !slugRe.MatchString(slug) {
		return nil, utils.ErrBadRequest("slug must be lowercase and dash-separated (e.g. color-match)")
	}
	if req.AgeCategoryID < 1 {
		return nil, utils.ErrBadRequest("age_category_id must be >= 1")
	}
	if createdBy <= 0 {
		return nil, utils.ErrInternal()
	}

	ok, err := s.gameRepo.AgeCategoryExists(ctx, req.AgeCategoryID)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if !ok {
		return nil, utils.ErrBadRequest("age_category_id not found")
	}

	exists, err := s.gameRepo.SlugExists(ctx, slug, nil)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if exists {
		return nil, utils.ErrBadRequest("slug already exists")
	}

	free := true
	if req.Free != nil {
		free = *req.Free
	}

	var thumb sql.NullString
	if strings.TrimSpace(req.Thumbnail) != "" {
		thumb = sql.NullString{String: strings.TrimSpace(req.Thumbnail), Valid: true}
	}
	var gameURL sql.NullString
	if strings.TrimSpace(req.GameURL) != "" {
		gameURL = sql.NullString{String: strings.TrimSpace(req.GameURL), Valid: true}
	}

	g, err := s.gameRepo.CreateAdminGame(ctx, repos.CreateAdminGameInput{
		Title:         title,
		Slug:          slug,
		Thumbnail:     thumb,
		GameURL:       gameURL,
		AgeCategoryID: req.AgeCategoryID,
		Free:          free,
		CreatedBy:     createdBy,
	})
	if err != nil {
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) UpdateAdminGame(ctx context.Context, id int64, req models.UpdateGameRequest) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	_, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, utils.ErrBadRequest("title cannot be empty")
		}
		if len(t) > 150 {
			return nil, utils.ErrBadRequest("title must be <= 150 chars")
		}
		req.Title = &t
	}

	if req.Slug != nil {
		slug := strings.TrimSpace(*req.Slug)
		if slug == "" {
			return nil, utils.ErrBadRequest("slug cannot be empty")
		}
		if len(slug) > 150 {
			return nil, utils.ErrBadRequest("slug must be <= 150 chars")
		}
		if !slugRe.MatchString(slug) {
			return nil, utils.ErrBadRequest("slug must be lowercase and dash-separated (e.g. color-match)")
		}

		exists, err := s.gameRepo.SlugExists(ctx, slug, &id)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		if exists {
			return nil, utils.ErrBadRequest("slug already exists")
		}
		req.Slug = &slug
	}

	if req.AgeCategoryID != nil {
		if *req.AgeCategoryID < 1 {
			return nil, utils.ErrBadRequest("age_category_id must be >= 1")
		}
		ok, err := s.gameRepo.AgeCategoryExists(ctx, *req.AgeCategoryID)
		if err != nil {
			return nil, utils.ErrInternal()
		}
		if !ok {
			return nil, utils.ErrBadRequest("age_category_id not found")
		}
	}

	g, err := s.gameRepo.UpdateAdminGame(ctx, id, repos.UpdateAdminGameInput{
		Title:         req.Title,
		Slug:          req.Slug,
		Thumbnail:     req.Thumbnail,
		GameURL:       req.GameURL,
		AgeCategoryID: req.AgeCategoryID,
		Free:          req.Free,
	})
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) PublishAdminGame(ctx context.Context, id int64) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	g0, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}
	if g0.Status == "archived" {
		return nil, utils.ErrBadRequest("archived game cannot be published")
	}

	g, err := s.gameRepo.SetStatus(ctx, id, "active")
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func (s *GameService) UnpublishAdminGame(ctx context.Context, id int64) (*models.AdminGameDTO, error) {
	if id < 1 {
		return nil, utils.ErrBadRequest("id must be an integer >= 1")
	}

	g0, err := s.gameRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}
	if g0.Status == "archived" {
		return nil, utils.ErrBadRequest("archived game cannot be unpublished")
	}

	g, err := s.gameRepo.SetStatus(ctx, id, "draft")
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return nil, utils.ErrNotFound("game not found")
		}
		return nil, utils.ErrInternal()
	}

	dto := toAdminGameDTO(*g)
	return &dto, nil
}

func toAdminGameDTO(g repos.Game) models.AdminGameDTO {
	var thumb string
	if g.Thumbnail.Valid {
		thumb = g.Thumbnail.String
	}
	var url string
	if g.GameURL.Valid {
		url = g.GameURL.String
	}

	return models.AdminGameDTO{
		ID:            g.ID,
		Title:         g.Title,
		Slug:          g.Slug,
		Status:        models.GameStatus(g.Status),
		Thumbnail:     thumb,
		GameURL:       url,
		AgeCategoryID: g.AgeCategoryID,
		Free:          g.Free,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
	}
}
